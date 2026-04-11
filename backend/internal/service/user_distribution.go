package service

import (
	"context"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrDistributionInviteCodeRequired        = infraerrors.BadRequest("DISTRIBUTION_INVITE_CODE_REQUIRED", "invite code is required")
	ErrDistributionInviteCodeInvalid         = infraerrors.NotFound("DISTRIBUTION_INVITE_CODE_INVALID", "invite code is invalid")
	ErrDistributionInviterAlreadyBound       = infraerrors.Conflict("DISTRIBUTION_INVITER_ALREADY_BOUND", "inviter already bound")
	ErrDistributionCannotBindSelf            = infraerrors.BadRequest("DISTRIBUTION_CANNOT_BIND_SELF", "cannot bind yourself as inviter")
	ErrDistributionCommissionRateInvalid     = infraerrors.BadRequest("DISTRIBUTION_COMMISSION_RATE_INVALID", "commission rate must be between 0 and 1")
	ErrDistributionUnsupportedRepo           = infraerrors.InternalServer("DISTRIBUTION_REPOSITORY_UNSUPPORTED", "distribution repository capability is unavailable")
	ErrDistributionWithdrawalAmountInvalid   = infraerrors.BadRequest("DISTRIBUTION_WITHDRAWAL_AMOUNT_INVALID", "withdrawal amount must be greater than 0")
	ErrDistributionWithdrawalAccountRequired = infraerrors.BadRequest("DISTRIBUTION_WITHDRAWAL_ACCOUNT_REQUIRED", "withdrawal account is required")
	ErrDistributionWithdrawalInsufficient    = infraerrors.BadRequest("DISTRIBUTION_WITHDRAWAL_INSUFFICIENT", "insufficient available commission")
	ErrDistributionWithdrawalPendingExists   = infraerrors.Conflict("DISTRIBUTION_WITHDRAWAL_PENDING_EXISTS", "an existing pending withdrawal request must be reviewed first")
	ErrDistributionWithdrawalNotFound        = infraerrors.NotFound("DISTRIBUTION_WITHDRAWAL_NOT_FOUND", "withdrawal request not found")
	ErrDistributionWithdrawalNotPending      = infraerrors.Conflict("DISTRIBUTION_WITHDRAWAL_NOT_PENDING", "withdrawal request is not pending")
	ErrDistributionWithdrawalStatusInvalid   = infraerrors.BadRequest("DISTRIBUTION_WITHDRAWAL_STATUS_INVALID", "withdrawal status must be approved or rejected")
)

type DistributionProfile struct {
	UserID                   int64      `json:"user_id"`
	InviterUserID            *int64     `json:"inviter_user_id"`
	InviteCode               string     `json:"invite_code"`
	CommissionRate           float64    `json:"commission_rate"`
	TotalReferrals           int64      `json:"total_referrals"`
	TotalCommissionEarned    float64    `json:"total_commission_earned"`
	TotalReferralContribution float64   `json:"total_referral_contribution"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type DistributionReferral struct {
	UserID            int64     `json:"user_id"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	BoundAt           time.Time `json:"bound_at"`
	TotalContribution float64   `json:"total_contribution"`
}

type DistributionCommissionRecord struct {
	ID               int64     `json:"id"`
	InviterUserID    int64     `json:"inviter_user_id"`
	InviteeUserID    int64     `json:"invitee_user_id"`
	InviteeEmail     string    `json:"invitee_email"`
	InviteeUsername  string    `json:"invitee_username"`
	TopupAmount      float64   `json:"topup_amount"`
	CommissionRate   float64   `json:"commission_rate"`
	CommissionAmount float64   `json:"commission_amount"`
	CommissionLevel  int       `json:"commission_level"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type DistributionTeamMember struct {
	UserID               int64     `json:"user_id"`
	Email                string    `json:"email"`
	Username             string    `json:"username"`
	BoundAt              time.Time `json:"bound_at"`
	TotalContribution    float64   `json:"total_contribution"`
	CommissionGenerated  float64   `json:"commission_generated"`
	TeamLevel            int       `json:"team_level"`
}

type DistributionSourceStat struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

type DistributionSummary struct {
	UserID                   int64                    `json:"user_id"`
	InviteCode               string                   `json:"invite_code"`
	TotalCommissionEarned    float64                  `json:"total_commission_earned"`
	TotalCommissionWithdrawn float64                  `json:"total_commission_withdrawn"`
	PendingWithdrawalAmount  float64                  `json:"pending_withdrawal_amount"`
	AvailableCommission      float64                  `json:"available_commission"`
	ThisMonthCommission      float64                  `json:"this_month_commission"`
	Level1TeamCount          int64                    `json:"level1_team_count"`
	Level2TeamCount          int64                    `json:"level2_team_count"`
	TotalTeamContribution    float64                  `json:"total_team_contribution"`
	SourceStats              []DistributionSourceStat `json:"source_stats,omitempty"`
}

type DistributionWithdrawalRequest struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Amount         float64    `json:"amount"`
	AccountType    string     `json:"account_type"`
	AccountRef     string     `json:"account_ref"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes,omitempty"`
	ReviewNote     string     `json:"review_note,omitempty"`
	ReviewedByUserID *int64   `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type DistributionCommissionBreakdown struct {
	InviterUserID    int64   `json:"inviter_user_id"`
	CommissionAmount float64 `json:"commission_amount"`
	CommissionLevel  int     `json:"commission_level"`
}

type DistributionCommissionResult struct {
	InviterUserID               int64                           `json:"inviter_user_id"`
	CommissionAmount            float64                         `json:"commission_amount"`
	SecondLevelInviterUserID    int64                           `json:"second_level_inviter_user_id"`
	SecondLevelCommissionAmount float64                         `json:"second_level_commission_amount"`
	TotalCommissionAmount       float64                         `json:"total_commission_amount"`
	Breakdown                   []DistributionCommissionBreakdown `json:"breakdown"`
}

type UserDistributionRepository interface {
	GetDistributionProfile(ctx context.Context, userID int64) (*DistributionProfile, error)
	GetDistributionSummary(ctx context.Context, userID int64) (*DistributionSummary, error)
	BindInviterByInviteCode(ctx context.Context, userID int64, inviteCode string) error
	ListDistributionReferrals(ctx context.Context, userID int64, params pagination.PaginationParams) ([]DistributionReferral, *pagination.PaginationResult, error)
	ListDistributionTeam(ctx context.Context, userID int64, params pagination.PaginationParams, level int) ([]DistributionTeamMember, *pagination.PaginationResult, error)
	ListDistributionCommissions(ctx context.Context, inviterUserID int64, params pagination.PaginationParams, level int) ([]DistributionCommissionRecord, *pagination.PaginationResult, error)
	ListDistributionSourceStats(ctx context.Context, inviterUserID int64) ([]DistributionSourceStat, error)
	CreateDistributionWithdrawalRequest(ctx context.Context, userID int64, amount float64, accountType, accountRef, notes string) (*DistributionWithdrawalRequest, error)
	UpsertDistributionInviteAttribution(ctx context.Context, inviteeUserID int64, inviteCode, source string) error
	ListDistributionWithdrawalRequests(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]DistributionWithdrawalRequest, *pagination.PaginationResult, error)
	ReviewDistributionWithdrawalRequest(ctx context.Context, userID, withdrawalID int64, status, reviewNote string, reviewerUserID int64) (*DistributionWithdrawalRequest, error)
	SetDistributionCommissionRate(ctx context.Context, userID int64, rate float64) error
	ApplyTopupDistributionCommission(ctx context.Context, inviteeUserID int64, topupAmount float64, notes string) (*DistributionCommissionResult, error)
}

func (s *UserService) distributionRepo() (UserDistributionRepository, error) {
	repo, ok := s.userRepo.(UserDistributionRepository)
	if !ok {
		return nil, ErrDistributionUnsupportedRepo
	}
	return repo, nil
}

func (s *UserService) GetDistributionProfile(ctx context.Context, userID int64) (*DistributionProfile, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, err
	}
	profile, err := repo.GetDistributionProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get distribution profile: %w", err)
	}
	return profile, nil
}

func (s *UserService) GetDistributionSummary(ctx context.Context, userID int64) (*DistributionSummary, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, err
	}
	summary, err := repo.GetDistributionSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get distribution summary: %w", err)
	}
	sourceStats, err := repo.ListDistributionSourceStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list distribution source stats: %w", err)
	}
	summary.SourceStats = sourceStats
	return summary, nil
}

func (s *UserService) BindInviterByInviteCode(ctx context.Context, userID int64, inviteCode string) error {
	if inviteCode == "" {
		return ErrDistributionInviteCodeRequired
	}
	repo, err := s.distributionRepo()
	if err != nil {
		return err
	}
	if err := repo.BindInviterByInviteCode(ctx, userID, inviteCode); err != nil {
		return fmt.Errorf("bind inviter by invite code: %w", err)
	}
	return nil
}

func (s *UserService) UpsertDistributionInviteAttribution(ctx context.Context, inviteeUserID int64, inviteCode, source string) error {
	if inviteCode == "" {
		return nil
	}
	repo, err := s.distributionRepo()
	if err != nil {
		return err
	}
	if err := repo.UpsertDistributionInviteAttribution(ctx, inviteeUserID, inviteCode, source); err != nil {
		return fmt.Errorf("upsert distribution invite attribution: %w", err)
	}
	return nil
}

func (s *UserService) ListDistributionReferrals(ctx context.Context, userID int64, params pagination.PaginationParams) ([]DistributionReferral, *pagination.PaginationResult, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, nil, err
	}
	list, result, err := repo.ListDistributionReferrals(ctx, userID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution referrals: %w", err)
	}
	return list, result, nil
}

func (s *UserService) ListDistributionTeam(ctx context.Context, userID int64, params pagination.PaginationParams, level int) ([]DistributionTeamMember, *pagination.PaginationResult, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, nil, err
	}
	list, result, err := repo.ListDistributionTeam(ctx, userID, params, level)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution team: %w", err)
	}
	return list, result, nil
}

func (s *UserService) ListDistributionCommissions(ctx context.Context, inviterUserID int64, params pagination.PaginationParams, level int) ([]DistributionCommissionRecord, *pagination.PaginationResult, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, nil, err
	}
	list, result, err := repo.ListDistributionCommissions(ctx, inviterUserID, params, level)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution commissions: %w", err)
	}
	return list, result, nil
}

func (s *UserService) CreateDistributionWithdrawalRequest(ctx context.Context, userID int64, amount float64, accountType, accountRef, notes string) (*DistributionWithdrawalRequest, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, err
	}
	item, err := repo.CreateDistributionWithdrawalRequest(ctx, userID, amount, accountType, accountRef, notes)
	if err != nil {
		return nil, fmt.Errorf("create distribution withdrawal request: %w", err)
	}
	return item, nil
}

func (s *UserService) ListDistributionWithdrawalRequests(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]DistributionWithdrawalRequest, *pagination.PaginationResult, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, nil, err
	}
	items, result, err := repo.ListDistributionWithdrawalRequests(ctx, userID, params, status)
	if err != nil {
		return nil, nil, fmt.Errorf("list distribution withdrawal requests: %w", err)
	}
	return items, result, nil
}

func (s *UserService) ReviewDistributionWithdrawalRequest(ctx context.Context, userID, withdrawalID int64, status, reviewNote string, reviewerUserID int64) (*DistributionWithdrawalRequest, error) {
	repo, err := s.distributionRepo()
	if err != nil {
		return nil, err
	}
	item, err := repo.ReviewDistributionWithdrawalRequest(ctx, userID, withdrawalID, status, reviewNote, reviewerUserID)
	if err != nil {
		return nil, fmt.Errorf("review distribution withdrawal request: %w", err)
	}
	return item, nil
}
