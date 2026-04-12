package handler

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ChangePasswordRequest represents the change password request payload
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	Username *string `json:"username"`
}

// GetProfile handles getting user profile
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userData, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserFromService(userData))
}

// ChangePassword handles changing user password
// POST /api/v1/users/me/password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.ChangePasswordRequest{
		CurrentPassword: req.OldPassword,
		NewPassword:     req.NewPassword,
	}
	err := h.userService.ChangePassword(c.Request.Context(), subject.UserID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Password changed successfully"})
}

// UpdateProfile handles updating user profile
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.UpdateProfileRequest{
		Username: req.Username,
	}
	updatedUser, err := h.userService.UpdateProfile(c.Request.Context(), subject.UserID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserFromService(updatedUser))
}

type BindDistributionInviterRequest struct {
	InviteCode string `json:"invite_code"`
}

// GetDistributionProfile handles getting current user's distribution profile
// GET /api/v1/user/distribution/profile
func (h *UserHandler) GetDistributionProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	profile, err := h.userService.GetDistributionProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, profile)
}

// GetDistributionSummary handles getting current user's distribution dashboard summary
// GET /api/v1/user/distribution/summary
func (h *UserHandler) GetDistributionSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	summary, err := h.userService.GetDistributionSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, summary)
}

// BindDistributionInviter handles binding inviter using invite code
// POST /api/v1/user/distribution/bind
func (h *UserHandler) BindDistributionInviter(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req BindDistributionInviterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.userService.BindInviterByInviteCode(c.Request.Context(), subject.UserID, strings.TrimSpace(req.InviteCode)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "inviter bound"})
}

// ListDistributionReferrals handles listing direct referrals
// GET /api/v1/user/distribution/referrals
func (h *UserHandler) ListDistributionReferrals(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	items, result, err := h.userService.ListDistributionReferrals(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, items, result.Total, page, pageSize)
}

// ListDistributionTeam handles listing current user's level-1 or level-2 team
// GET /api/v1/user/distribution/team?level=1|2
func (h *UserHandler) ListDistributionTeam(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	rawLevel := strings.TrimSpace(c.Query("level"))
	parsed, err := strconv.Atoi(rawLevel)
	if err != nil || (parsed != 1 && parsed != 2) {
		response.BadRequest(c, "Invalid level, expected 1 or 2")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	items, result, err := h.userService.ListDistributionTeam(c.Request.Context(), subject.UserID, params, parsed)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, items, result.Total, page, pageSize)
}

// ListDistributionCommissions handles listing current user's commission records
// GET /api/v1/user/distribution/commissions
func (h *UserHandler) ListDistributionCommissions(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	level := 0
	if rawLevel := strings.TrimSpace(c.Query("level")); rawLevel != "" {
		parsed, err := strconv.Atoi(rawLevel)
		if err != nil {
			response.BadRequest(c, "Invalid level, expected 1 or 2")
			return
		}
		if parsed != 1 && parsed != 2 {
			response.BadRequest(c, "Invalid level, expected 1 or 2")
			return
		}
		level = parsed
	}

	items, result, err := h.userService.ListDistributionCommissions(c.Request.Context(), subject.UserID, params, level)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, items, result.Total, page, pageSize)
}

type CreateDistributionWithdrawalRequest struct {
	Amount      float64 `json:"amount"`
	AccountType string  `json:"account_type"`
	AccountRef  string  `json:"account_ref"`
	Notes       string  `json:"notes"`
}

// CreateDistributionWithdrawal handles creating a withdrawal request
// POST /api/v1/user/distribution/withdrawals
func (h *UserHandler) CreateDistributionWithdrawal(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateDistributionWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	item, err := h.userService.CreateDistributionWithdrawalRequest(c.Request.Context(), subject.UserID, req.Amount, req.AccountType, req.AccountRef, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, item)
}

// ListDistributionWithdrawals handles listing withdrawal requests
// GET /api/v1/user/distribution/withdrawals
func (h *UserHandler) ListDistributionWithdrawals(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	status := strings.TrimSpace(c.Query("status"))

	items, result, err := h.userService.ListDistributionWithdrawalRequests(c.Request.Context(), subject.UserID, params, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, items, result.Total, page, pageSize)
}
