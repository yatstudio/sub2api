package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"math"
	"strings"

	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const (
	defaultDistributionCommissionRate = 0.10
	inviteCodeBytes                  = 6

	distributionWithdrawalStatusPending  = "pending"
	distributionWithdrawalStatusApproved = "approved"
	distributionWithdrawalStatusRejected = "rejected"

	minDistributionWithdrawalAmount = 10.0
	distributionWithdrawalDailyLimit = 1
)

func (r *userRepository) GetDistributionProfile(ctx context.Context, userID int64) (*service.DistributionProfile, error) {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, err
	}

	var profile service.DistributionProfile
	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			user_id,
			inviter_user_id,
			invite_code,
			commission_rate,
			total_referrals,
			total_commission_earned,
			total_contribution,
			created_at,
			updated_at
		FROM user_distributions
		WHERE user_id = $1
	`, []any{userID},
		&profile.UserID,
		&profile.InviterUserID,
		&profile.InviteCode,
		&profile.CommissionRate,
		&profile.TotalReferrals,
		&profile.TotalCommissionEarned,
		&profile.TotalReferralContribution,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}

	return &profile, nil
}

func (r *userRepository) GetDistributionSummary(ctx context.Context, userID int64) (*service.DistributionSummary, error) {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, err
	}

	var summary service.DistributionSummary
	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			d.user_id,
			d.invite_code,
			d.total_commission_earned,
			d.total_commission_withdrawn,
			COALESCE((
				SELECT SUM(w.amount)
				FROM distribution_withdrawal_requests w
				WHERE w.user_id = d.user_id
				  AND w.status = 'pending'
			), 0)::DECIMAL(20,8) AS pending_withdrawal_amount,
			(d.total_commission_earned - d.total_commission_withdrawn - COALESCE((
				SELECT SUM(w.amount)
				FROM distribution_withdrawal_requests w
				WHERE w.user_id = d.user_id
				  AND w.status = 'pending'
			), 0)::DECIMAL(20,8)) AS available_commission,
			COALESCE((
				SELECT SUM(dc.commission_amount)
				FROM distribution_commissions dc
				WHERE dc.inviter_user_id = d.user_id
				  AND dc.created_at >= date_trunc('month', NOW())
			), 0)::DECIMAL(20,8) AS this_month_commission,
			COALESCE((
				SELECT COUNT(1)
				FROM user_distributions l1
				JOIN users u1 ON u1.id = l1.user_id
				WHERE l1.inviter_user_id = d.user_id
				  AND u1.deleted_at IS NULL
			), 0)::BIGINT AS level1_team_count,
			COALESCE((
				SELECT COUNT(1)
				FROM user_distributions l2
				JOIN user_distributions l1 ON l1.user_id = l2.inviter_user_id
				JOIN users u2 ON u2.id = l2.user_id
				WHERE l1.inviter_user_id = d.user_id
				  AND u2.deleted_at IS NULL
			), 0)::BIGINT AS level2_team_count,
			COALESCE((
				SELECT SUM(member.total_contribution)
				FROM user_distributions member
				JOIN users um ON um.id = member.user_id
				WHERE um.deleted_at IS NULL
				  AND (
					member.inviter_user_id = d.user_id
					OR member.inviter_user_id IN (
						SELECT l1.user_id FROM user_distributions l1 WHERE l1.inviter_user_id = d.user_id
					)
				  )
			), 0)::DECIMAL(20,8) AS total_team_contribution
		FROM user_distributions d
		WHERE d.user_id = $1
	`, []any{userID},
		&summary.UserID,
		&summary.InviteCode,
		&summary.TotalCommissionEarned,
		&summary.TotalCommissionWithdrawn,
		&summary.PendingWithdrawalAmount,
		&summary.AvailableCommission,
		&summary.ThisMonthCommission,
		&summary.Level1TeamCount,
		&summary.Level2TeamCount,
		&summary.TotalTeamContribution,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}

	return &summary, nil
}

func (r *userRepository) GetDistributionOverview(ctx context.Context) (*service.DistributionOverview, error) {
	overview := &service.DistributionOverview{}

	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(1)
		FROM user_distributions
	`, nil, &overview.TotalDistributors); err != nil {
		return nil, err
	}

	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(1)
		FROM user_distributions
		WHERE inviter_user_id IS NOT NULL
	`, nil, &overview.TotalBoundUsers); err != nil {
		return nil, err
	}

	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(1), COALESCE(SUM(amount), 0)::DECIMAL(20,8)
		FROM distribution_withdrawals
		WHERE status = $1
	`, []any{distributionWithdrawalStatusPending}, &overview.PendingWithdrawalCount, &overview.PendingWithdrawalAmount); err != nil {
		return nil, err
	}

	sourceStats, err := r.ListDistributionSourceStats(ctx, 0)
	if err != nil {
		return nil, err
	}
	overview.SourceStats = sourceStats

	return overview, nil
}

func (r *userRepository) ListDistributionSourceStats(ctx context.Context, inviterUserID int64) ([]service.DistributionSourceStat, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT COALESCE(source, 'direct') AS source, COUNT(1)
		FROM distribution_invite_attributions
		WHERE ($1 = 0 OR inviter_user_id = $1)
		GROUP BY COALESCE(source, 'direct')
		ORDER BY COUNT(1) DESC, source ASC
	`, inviterUserID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "distribution_invite_attributions") {
			return []service.DistributionSourceStat{}, nil
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.DistributionSourceStat, 0, 4)
	for rows.Next() {
		var item service.DistributionSourceStat
		if err := rows.Scan(&item.Source, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func normalizeDistributionSource(source string) string {
	s := strings.ToLower(strings.TrimSpace(source))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "direct"
	}

	switch s {
	case "direct", "wechat", "telegram", "discord", "x", "youtube", "website", "group", "campaign", "referral":
		return s
	default:
		return "direct"
	}
}

func (r *userRepository) UpsertDistributionInviteAttribution(ctx context.Context, inviteeUserID int64, inviteCode, source string) error {
	code := normalizeInviteCode(inviteCode)
	if code == "" {
		return nil
	}
	normalizedSource := normalizeDistributionSource(source)

	if _, err := r.sql.ExecContext(ctx, `
		INSERT INTO distribution_invite_attributions (invitee_user_id, inviter_user_id, invite_code, source)
		SELECT $1, d.user_id, d.invite_code, $3
		FROM user_distributions d
		WHERE d.invite_code = $2
		ON CONFLICT (invitee_user_id)
		DO UPDATE SET
			inviter_user_id = EXCLUDED.inviter_user_id,
			invite_code = EXCLUDED.invite_code,
			source = EXCLUDED.source,
			updated_at = NOW()
	`, inviteeUserID, code, normalizedSource); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "distribution_invite_attributions") {
			return nil
		}
		return err
	}

	return nil
}

func (r *userRepository) BindInviterByInviteCode(ctx context.Context, userID int64, inviteCode string) error {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return err
	}
	code := normalizeInviteCode(inviteCode)
	if code == "" {
		return service.ErrDistributionInviteCodeRequired
	}

	var inviterID int64
	err := scanSingleRow(ctx, r.sql, `
		WITH RECURSIVE inviter AS (
			SELECT user_id
			FROM user_distributions
			WHERE invite_code = $2
		), descendants AS (
			SELECT user_id
			FROM user_distributions
			WHERE inviter_user_id = $1
			UNION ALL
			SELECT d.user_id
			FROM user_distributions d
			JOIN descendants ds ON d.inviter_user_id = ds.user_id
		), updated AS (
			UPDATE user_distributions target
			SET inviter_user_id = (SELECT user_id FROM inviter),
				updated_at = NOW()
			WHERE target.user_id = $1
			  AND target.inviter_user_id IS NULL
			  AND (SELECT user_id FROM inviter) IS NOT NULL
			  AND (SELECT user_id FROM inviter) <> $1
			  AND (SELECT user_id FROM inviter) NOT IN (SELECT user_id FROM descendants)
			RETURNING (SELECT user_id FROM inviter) AS inviter_user_id
		), inc AS (
			UPDATE user_distributions
			SET total_referrals = total_referrals + 1,
				updated_at = NOW()
			WHERE user_id = (SELECT inviter_user_id FROM updated)
		)
		SELECT inviter_user_id FROM updated
	`, []any{userID, code}, &inviterID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var existingInviterID *int64
	if qErr := scanSingleRow(ctx, r.sql, `SELECT inviter_user_id FROM user_distributions WHERE user_id = $1`, []any{userID}, &existingInviterID); qErr == nil {
		if existingInviterID != nil {
			return service.ErrDistributionInviterAlreadyBound
		}
	}

	var codeOwnerID int64
	if qErr := scanSingleRow(ctx, r.sql, `SELECT user_id FROM user_distributions WHERE invite_code = $1`, []any{code}, &codeOwnerID); qErr != nil {
		if errors.Is(qErr, sql.ErrNoRows) {
			return service.ErrDistributionInviteCodeInvalid
		}
		return qErr
	}
	if codeOwnerID == userID {
		return service.ErrDistributionCannotBindSelf
	}

	return service.ErrDistributionInviterAlreadyBound
}

func (r *userRepository) ListDistributionReferrals(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.DistributionReferral, *pagination.PaginationResult, error) {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, nil, err
	}

	var total int64
	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(1)
		FROM user_distributions d
		JOIN users u ON u.id = d.user_id
		WHERE d.inviter_user_id = $1
		  AND u.deleted_at IS NULL
	`, []any{userID}, &total); err != nil {
		return nil, nil, err
	}

	rows, err := r.sql.QueryContext(ctx, `
		SELECT
			d.user_id,
			u.email,
			u.username,
			d.created_at,
			d.total_contribution
		FROM user_distributions d
		JOIN users u ON u.id = d.user_id
		WHERE d.inviter_user_id = $1
		  AND u.deleted_at IS NULL
		ORDER BY d.created_at DESC, d.user_id DESC
		LIMIT $2 OFFSET $3
	`, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.DistributionReferral, 0, params.Limit())
	for rows.Next() {
		var item service.DistributionReferral
		if err := rows.Scan(&item.UserID, &item.Email, &item.Username, &item.BoundAt, &item.TotalContribution); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *userRepository) ListDistributionTeam(ctx context.Context, userID int64, params pagination.PaginationParams, level int) ([]service.DistributionTeamMember, *pagination.PaginationResult, error) {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, nil, err
	}

	if level != 1 && level != 2 {
		return []service.DistributionTeamMember{}, paginationResultFromTotal(0, params), nil
	}

	var (
		countSQL string
		listSQL  string
		args     []any
	)
	if level == 1 {
		countSQL = `
			SELECT COUNT(1)
			FROM user_distributions d
			JOIN users u ON u.id = d.user_id
			WHERE d.inviter_user_id = $1
			  AND u.deleted_at IS NULL
		`
		listSQL = `
			SELECT
				d.user_id,
				COALESCE(u.email, ''),
				COALESCE(u.username, ''),
				d.created_at,
				d.total_contribution,
				COALESCE((
					SELECT SUM(dc.commission_amount)
					FROM distribution_commissions dc
					WHERE dc.inviter_user_id = $1
					  AND dc.invitee_user_id = d.user_id
					  AND dc.commission_level = 1
				), 0)::DECIMAL(20,8) AS commission_generated,
				1 AS team_level
			FROM user_distributions d
			JOIN users u ON u.id = d.user_id
			WHERE d.inviter_user_id = $1
			  AND u.deleted_at IS NULL
			ORDER BY d.created_at DESC, d.user_id DESC
			LIMIT $2 OFFSET $3
		`
		args = []any{userID}
	} else {
		countSQL = `
			SELECT COUNT(1)
			FROM user_distributions d2
			JOIN user_distributions d1 ON d1.user_id = d2.inviter_user_id
			JOIN users u ON u.id = d2.user_id
			WHERE d1.inviter_user_id = $1
			  AND u.deleted_at IS NULL
		`
		listSQL = `
			SELECT
				d2.user_id,
				COALESCE(u.email, ''),
				COALESCE(u.username, ''),
				d2.created_at,
				d2.total_contribution,
				COALESCE((
					SELECT SUM(dc.commission_amount)
					FROM distribution_commissions dc
					WHERE dc.inviter_user_id = $1
					  AND dc.invitee_user_id = d2.user_id
					  AND dc.commission_level = 2
				), 0)::DECIMAL(20,8) AS commission_generated,
				2 AS team_level
			FROM user_distributions d2
			JOIN user_distributions d1 ON d1.user_id = d2.inviter_user_id
			JOIN users u ON u.id = d2.user_id
			WHERE d1.inviter_user_id = $1
			  AND u.deleted_at IS NULL
			ORDER BY d2.created_at DESC, d2.user_id DESC
			LIMIT $2 OFFSET $3
		`
		args = []any{userID}
	}

	var total int64
	if err := scanSingleRow(ctx, r.sql, countSQL, args, &total); err != nil {
		return nil, nil, err
	}

	rows, err := r.sql.QueryContext(ctx, listSQL, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.DistributionTeamMember, 0, params.Limit())
	for rows.Next() {
		var item service.DistributionTeamMember
		if err := rows.Scan(
			&item.UserID,
			&item.Email,
			&item.Username,
			&item.BoundAt,
			&item.TotalContribution,
			&item.CommissionGenerated,
			&item.TeamLevel,
		); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *userRepository) ListDistributionCommissions(ctx context.Context, inviterUserID int64, params pagination.PaginationParams, level int) ([]service.DistributionCommissionRecord, *pagination.PaginationResult, error) {
	if err := r.ensureDistributionProfile(ctx, inviterUserID); err != nil {
		return nil, nil, err
	}

	levelFilter := 0
	if level == 1 || level == 2 {
		levelFilter = level
	}

	countSQL := `
		SELECT COUNT(1)
		FROM distribution_commissions dc
		WHERE dc.inviter_user_id = $1
	`
	countArgs := []any{inviterUserID}
	if levelFilter > 0 {
		countSQL += ` AND dc.commission_level = $2`
		countArgs = append(countArgs, levelFilter)
	}

	var total int64
	if err := scanSingleRow(ctx, r.sql, countSQL, countArgs, &total); err != nil {
		return nil, nil, err
	}

	listSQL := `
		SELECT
			dc.id,
			dc.inviter_user_id,
			dc.invitee_user_id,
			COALESCE(u.email, ''),
			COALESCE(u.username, ''),
			dc.topup_amount,
			dc.commission_rate,
			dc.commission_amount,
			dc.commission_level,
			COALESCE(dc.notes, ''),
			dc.created_at
		FROM distribution_commissions dc
		LEFT JOIN users u ON u.id = dc.invitee_user_id
		WHERE dc.inviter_user_id = $1
	`
	listArgs := []any{inviterUserID}
	if levelFilter > 0 {
		listSQL += ` AND dc.commission_level = $2`
		listArgs = append(listArgs, levelFilter)
		listSQL += ` ORDER BY dc.created_at DESC, dc.id DESC LIMIT $3 OFFSET $4`
		listArgs = append(listArgs, params.Limit(), params.Offset())
	} else {
		listSQL += ` ORDER BY dc.created_at DESC, dc.id DESC LIMIT $2 OFFSET $3`
		listArgs = append(listArgs, params.Limit(), params.Offset())
	}

	rows, err := r.sql.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.DistributionCommissionRecord, 0, params.Limit())
	for rows.Next() {
		var item service.DistributionCommissionRecord
		if err := rows.Scan(
			&item.ID,
			&item.InviterUserID,
			&item.InviteeUserID,
			&item.InviteeEmail,
			&item.InviteeUsername,
			&item.TopupAmount,
			&item.CommissionRate,
			&item.CommissionAmount,
			&item.CommissionLevel,
			&item.Notes,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *userRepository) CreateDistributionWithdrawalRequest(ctx context.Context, userID int64, amount float64, accountType, accountRef, notes string) (*service.DistributionWithdrawalRequest, error) {
	if amount <= 0 {
		return nil, service.ErrDistributionWithdrawalAmountInvalid
	}
	if amount < minDistributionWithdrawalAmount {
		return nil, service.ErrDistributionWithdrawalAmountTooSmall
	}
	if strings.TrimSpace(accountRef) == "" {
		return nil, service.ErrDistributionWithdrawalAccountRequired
	}
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, err
	}

	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		totalEarned    float64
		totalWithdrawn float64
		pendingAmount  float64
		pendingCount   int64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT
			total_commission_earned,
			total_commission_withdrawn,
			COALESCE((
				SELECT SUM(w.amount)
				FROM distribution_withdrawal_requests w
				WHERE w.user_id = $1
				  AND w.status = $2
			), 0)::DECIMAL(20,8) AS pending_amount,
			COALESCE((
				SELECT COUNT(1)
				FROM distribution_withdrawal_requests w
				WHERE w.user_id = $1
				  AND w.status = $2
			), 0) AS pending_count
		FROM user_distributions
		WHERE user_id = $1
		FOR UPDATE
	`, userID, distributionWithdrawalStatusPending).Scan(&totalEarned, &totalWithdrawn, &pendingAmount, &pendingCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}

	if pendingCount > 0 {
		return nil, service.ErrDistributionWithdrawalPendingExists
	}

	var dailyCount int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM distribution_withdrawal_requests
		WHERE user_id = $1
		  AND created_at >= NOW() - INTERVAL '24 hours'
	`, userID).Scan(&dailyCount); err != nil {
		return nil, err
	}
	if dailyCount >= distributionWithdrawalDailyLimit {
		return nil, service.ErrDistributionWithdrawalDailyLimit
	}

	available := totalEarned - totalWithdrawn - pendingAmount
	if amount > available {
		return nil, service.ErrDistributionWithdrawalInsufficient
	}

	var item service.DistributionWithdrawalRequest
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO distribution_withdrawal_requests (
			user_id,
			amount,
			account_type,
			account_ref,
			notes,
			status
		)
		VALUES ($1, $2::DECIMAL(20,8), NULLIF($3, ''), $4, NULLIF($5, ''), $6)
		RETURNING id, user_id, amount, COALESCE(account_type, ''), account_ref, status, COALESCE(notes, ''), created_at, updated_at
	`, userID, amount, strings.TrimSpace(accountType), strings.TrimSpace(accountRef), strings.TrimSpace(notes), distributionWithdrawalStatusPending).Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&item.AccountType,
		&item.AccountRef,
		&item.Status,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *userRepository) ListDistributionWithdrawalRequests(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]service.DistributionWithdrawalRequest, *pagination.PaginationResult, error) {
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, nil, err
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	if normalizedStatus != "" && normalizedStatus != distributionWithdrawalStatusPending && normalizedStatus != distributionWithdrawalStatusApproved && normalizedStatus != distributionWithdrawalStatusRejected {
		normalizedStatus = ""
	}

	countSQL := `SELECT COUNT(1) FROM distribution_withdrawal_requests WHERE user_id = $1`
	countArgs := []any{userID}
	if normalizedStatus != "" {
		countSQL += ` AND status = $2`
		countArgs = append(countArgs, normalizedStatus)
	}

	var total int64
	if err := scanSingleRow(ctx, r.sql, countSQL, countArgs, &total); err != nil {
		return nil, nil, err
	}

	listSQL := `
		SELECT
			id,
			user_id,
			amount,
			COALESCE(account_type, ''),
			account_ref,
			status,
			COALESCE(notes, ''),
			COALESCE(review_note, ''),
			reviewed_by_user_id,
			reviewed_at,
			created_at,
			updated_at
		FROM distribution_withdrawal_requests
		WHERE user_id = $1
	`
	listArgs := []any{userID}
	if normalizedStatus != "" {
		listSQL += ` AND status = $2 ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`
		listArgs = append(listArgs, normalizedStatus, params.Limit(), params.Offset())
	} else {
		listSQL += ` ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`
		listArgs = append(listArgs, params.Limit(), params.Offset())
	}

	rows, err := r.sql.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.DistributionWithdrawalRequest, 0, params.Limit())
	for rows.Next() {
		var item service.DistributionWithdrawalRequest
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Amount,
			&item.AccountType,
			&item.AccountRef,
			&item.Status,
			&item.Notes,
			&item.ReviewNote,
			&item.ReviewedByUserID,
			&item.ReviewedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *userRepository) ReviewDistributionWithdrawalRequest(ctx context.Context, userID, withdrawalID int64, status, reviewNote string, reviewerUserID int64) (*service.DistributionWithdrawalRequest, error) {
	nextStatus := strings.ToLower(strings.TrimSpace(status))
	if nextStatus != distributionWithdrawalStatusApproved && nextStatus != distributionWithdrawalStatusRejected {
		return nil, service.ErrDistributionWithdrawalStatusInvalid
	}
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return nil, err
	}

	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var amount float64
	var currentStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT amount, status
		FROM distribution_withdrawal_requests
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, withdrawalID, userID).Scan(&amount, &currentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDistributionWithdrawalNotFound
		}
		return nil, err
	}
	if currentStatus != distributionWithdrawalStatusPending {
		return nil, service.ErrDistributionWithdrawalNotPending
	}

	if nextStatus == distributionWithdrawalStatusApproved {
		res, err := tx.ExecContext(ctx, `
			UPDATE users
			SET balance = balance - $2::DECIMAL(20,8)
			WHERE id = $1
			  AND balance >= $2::DECIMAL(20,8)
		`, userID, amount)
		if err != nil {
			return nil, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil, service.ErrDistributionWithdrawalInsufficient
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE user_distributions
			SET total_commission_withdrawn = total_commission_withdrawn + $2::DECIMAL(20,8),
				updated_at = NOW()
			WHERE user_id = $1
		`, userID, amount); err != nil {
			return nil, err
		}
	}

	var item service.DistributionWithdrawalRequest
	if err := tx.QueryRowContext(ctx, `
		UPDATE distribution_withdrawal_requests
		SET status = $3,
			review_note = NULLIF($4, ''),
			reviewed_by_user_id = NULLIF($5, 0),
			reviewed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, amount, COALESCE(account_type, ''), account_ref, status, COALESCE(notes, ''), COALESCE(review_note, ''), reviewed_by_user_id, reviewed_at, created_at, updated_at
	`, withdrawalID, userID, nextStatus, strings.TrimSpace(reviewNote), reviewerUserID).Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&item.AccountType,
		&item.AccountRef,
		&item.Status,
		&item.Notes,
		&item.ReviewNote,
		&item.ReviewedByUserID,
		&item.ReviewedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *userRepository) SetDistributionCommissionRate(ctx context.Context, userID int64, rate float64) error {
	if rate < 0 || rate > 1 {
		return service.ErrDistributionCommissionRateInvalid
	}
	if err := r.ensureDistributionProfile(ctx, userID); err != nil {
		return err
	}

	res, err := r.sql.ExecContext(ctx, `
		UPDATE user_distributions
		SET commission_rate = $2,
			updated_at = NOW()
		WHERE user_id = $1
	`, userID, rate)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ApplyTopupDistributionCommission(ctx context.Context, inviteeUserID int64, topupAmount float64, notes string) (*service.DistributionCommissionResult, error) {
	if topupAmount <= 0 {
		return nil, nil
	}
	if err := r.ensureDistributionProfile(ctx, inviteeUserID); err != nil {
		return nil, err
	}

	var (
		level1InviterID sql.NullInt64
		level1Rate      sql.NullFloat64
		level2InviterID sql.NullInt64
		level2Rate      sql.NullFloat64
	)
	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			invitee.inviter_user_id AS level1_inviter_user_id,
			level1.commission_rate AS level1_rate,
			level1.inviter_user_id AS level2_inviter_user_id,
			level2.commission_rate AS level2_rate
		FROM user_distributions invitee
		LEFT JOIN user_distributions level1 ON level1.user_id = invitee.inviter_user_id
		LEFT JOIN user_distributions level2 ON level2.user_id = level1.inviter_user_id
		WHERE invitee.user_id = $1
	`, []any{inviteeUserID}, &level1InviterID, &level1Rate, &level2InviterID, &level2Rate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	round8 := func(v float64) float64 {
		return math.Round(v*1e8) / 1e8
	}

	payouts := make([]service.DistributionCommissionBreakdown, 0, 2)
	if level1InviterID.Valid && level1InviterID.Int64 > 0 && level1Rate.Valid && level1Rate.Float64 > 0 {
		amount := round8(topupAmount * level1Rate.Float64)
		if amount > 0 {
			payouts = append(payouts, service.DistributionCommissionBreakdown{
				InviterUserID:    level1InviterID.Int64,
				CommissionAmount: amount,
				CommissionLevel:  1,
			})
		}
	}
	if level2InviterID.Valid && level2InviterID.Int64 > 0 && level2Rate.Valid && level2Rate.Float64 > 0 {
		amount := round8(topupAmount * level2Rate.Float64)
		if amount > 0 {
			payouts = append(payouts, service.DistributionCommissionBreakdown{
				InviterUserID:    level2InviterID.Int64,
				CommissionAmount: amount,
				CommissionLevel:  2,
			})
		}
	}

	if len(payouts) == 0 {
		return nil, nil
	}

	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		UPDATE user_distributions
		SET total_contribution = total_contribution + $2::DECIMAL(20,8),
			updated_at = NOW()
		WHERE user_id = $1
	`, inviteeUserID, topupAmount); err != nil {
		return nil, err
	}

	normalizedNotes := strings.TrimSpace(notes)
	for _, payout := range payouts {
		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET balance = balance + $2::DECIMAL(20,8)
			WHERE id = $1
		`, payout.InviterUserID, payout.CommissionAmount); err != nil {
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE user_distributions
			SET total_commission_earned = total_commission_earned + $2::DECIMAL(20,8),
				updated_at = NOW()
			WHERE user_id = $1
		`, payout.InviterUserID, payout.CommissionAmount); err != nil {
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO distribution_commissions (
				inviter_user_id,
				invitee_user_id,
				topup_amount,
				commission_rate,
				commission_amount,
				commission_level,
				notes
			)
			VALUES ($1, $2, $3::DECIMAL(20,8), $4::DECIMAL(6,4), $5::DECIMAL(20,8), $6, NULLIF($7, ''))
		`, payout.InviterUserID, inviteeUserID, topupAmount, payout.CommissionAmount/topupAmount, payout.CommissionAmount, payout.CommissionLevel, normalizedNotes); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	result := &service.DistributionCommissionResult{Breakdown: payouts}
	for _, payout := range payouts {
		result.TotalCommissionAmount += payout.CommissionAmount
		switch payout.CommissionLevel {
		case 1:
			result.InviterUserID = payout.InviterUserID
			result.CommissionAmount = payout.CommissionAmount
		case 2:
			result.SecondLevelInviterUserID = payout.InviterUserID
			result.SecondLevelCommissionAmount = payout.CommissionAmount
		}
	}

	return result, nil
}

func (r *userRepository) ensureDistributionProfile(ctx context.Context, userID int64) error {
	exists, err := r.client.User.Query().Where(dbuser.IDEQ(userID)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return service.ErrUserNotFound
	}

	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		code, genErr := generateInviteCode()
		if genErr != nil {
			return genErr
		}
		_, err = r.sql.ExecContext(ctx, `
			INSERT INTO user_distributions (user_id, invite_code, commission_rate)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO NOTHING
		`, userID, code, defaultDistributionCommissionRate)
		if err == nil {
			return nil
		}
		if isUniqueViolation(err) {
			continue
		}
		return err
	}
	return fmt.Errorf("failed to allocate unique invite code")
}

func generateInviteCode() (string, error) {
	buf := make([]byte, inviteCodeBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// 6 bytes => 10 chars base32 without padding.
	code := strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), "=")
	return normalizeInviteCode(code), nil
}

func normalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
