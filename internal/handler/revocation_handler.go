package handler

import (
	"net/http"

	"auth-service/internal/service"
	"auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RevocationHandler handles token revocation endpoints
type RevocationHandler struct {
	revocationSvc service.RevocationService
}

// NewRevocationHandler creates a new revocation handler
func NewRevocationHandler(revocationSvc service.RevocationService) *RevocationHandler {
	return &RevocationHandler{
		revocationSvc: revocationSvc,
	}
}

// RevokeToken godoc
// @Summary Revoke a specific token
// @Description Adds a token to the denylist, preventing its use
// @Tags revocation
// @Accept json
// @Produce json
// @Param request body RevokeTokenRequest true "Token to revoke"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/revocation/token [post]
func (h *RevocationHandler) RevokeToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn(ctx).
			Err(err).
			Msg("Invalid revoke token request")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	if err := h.revocationSvc.RevokeToken(ctx, req.Token); err != nil {
		logger.Error(ctx).
			Err(err).
			Msg("Failed to revoke token")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke token: " + err.Error()})
		return
	}

	logger.Info(ctx).Msg("Token revoked successfully")
	c.JSON(http.StatusOK, MessageResponse{Message: "token revoked successfully"})
}

// RevokeUserSessions godoc
// @Summary Revoke all sessions for a user
// @Description Revokes all active sessions for a specific user across all organizations
// @Tags revocation
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/revocation/user/{user_id}/sessions [delete]
func (h *RevocationHandler) RevokeUserSessions(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Warn(ctx).
			Str("user_id_param", userIDStr).
			Msg("Invalid user ID format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
		return
	}

	if err := h.revocationSvc.RevokeUserSessions(ctx, userID); err != nil {
		logger.Error(ctx).
			Err(err).
			Str("target_user_id", userID.String()).
			Msg("Failed to revoke user sessions")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke user sessions: " + err.Error()})
		return
	}

	logger.Info(ctx).
		Str("target_user_id", userID.String()).
		Msg("User sessions revoked successfully")
	c.JSON(http.StatusOK, MessageResponse{Message: "user sessions revoked successfully"})
}

// RevokeOrgSessions godoc
// @Summary Revoke all sessions for an organization
// @Description Revokes all active sessions for all users in a specific organization
// @Tags revocation
// @Accept json
// @Produce json
// @Param org_id path string true "Organization ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/revocation/org/{org_id}/sessions [delete]
func (h *RevocationHandler) RevokeOrgSessions(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid organization ID"})
		return
	}

	if err := h.revocationSvc.RevokeOrgSessions(c.Request.Context(), orgID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke organization sessions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "organization sessions revoked successfully"})
}

// RevokeUserInOrg godoc
// @Summary Revoke user sessions in a specific organization
// @Description Revokes all sessions for a user in a specific organization only
// @Tags revocation
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param org_id path string true "Organization ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/revocation/user/{user_id}/org/{org_id}/sessions [delete]
func (h *RevocationHandler) RevokeUserInOrg(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
		return
	}

	orgIDStr := c.Param("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid organization ID"})
		return
	}

	if err := h.revocationSvc.RevokeUserInOrg(c.Request.Context(), userID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke user sessions in organization: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "user sessions in organization revoked successfully"})
}

// Request/Response types

type RevokeTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
