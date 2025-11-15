package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// OrganizationHandler handles organization-related endpoints
type OrganizationHandler struct {
	authService service.AuthService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(authService service.AuthService) *OrganizationHandler {
	return &OrganizationHandler{
		authService: authService,
	}
}

// CreateOrganization handles organization creation
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.OrganizationService().CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Organization created successfully",
	})
}

// GetOrganization handles getting organization details
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	response, err := h.authService.OrganizationService().GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateOrganization handles organization updates
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	var req service.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.OrganizationService().UpdateOrganization(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Organization updated successfully",
	})
}

// DeleteOrganization handles organization deletion
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	err := h.authService.OrganizationService().DeleteOrganization(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Organization deleted successfully",
	})
}

// ListUserOrganizations handles listing user's organizations
func (h *OrganizationHandler) ListUserOrganizations(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	response, err := h.authService.OrganizationService().ListUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// InviteUser handles inviting a user to an organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID := c.Param("orgId")

	var req service.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	// Set organization ID from URL param
	req.OrganizationID = orgID

	response, err := h.authService.OrganizationService().InviteUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Invitation sent successfully",
	})
}

// AcceptInvitation handles accepting an organization invitation
func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	userID, _ := c.Request.Context().Value("user_id").(string)

	response, err := h.authService.OrganizationService().AcceptInvitation(c.Request.Context(), token, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Invitation accepted successfully",
	})
}

// ListOrganizationMembers handles listing organization members
func (h *OrganizationHandler) ListOrganizationMembers(c *gin.Context) {
	orgID := c.Param("orgId")

	response, err := h.authService.OrganizationService().ListMembers(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateMembership handles updating a user's membership in an organization
func (h *OrganizationHandler) UpdateMembership(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	var req service.UpdateMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	_, err := h.authService.OrganizationService().UpdateMembership(c.Request.Context(), orgID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Membership updated successfully",
	})
}

// RemoveMember handles removing a member from an organization
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	err := h.authService.OrganizationService().RemoveMember(c.Request.Context(), orgID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Member removed successfully",
	})
}

// GetOrganizationInvitations handles getting pending invitations for an organization
func (h *OrganizationHandler) GetOrganizationInvitations(c *gin.Context) {
	orgID := c.Param("orgId")

	response, err := h.authService.OrganizationService().ListPendingInvitations(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CancelInvitation handles canceling an organization invitation
func (h *OrganizationHandler) CancelInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	err := h.authService.OrganizationService().CancelInvitation(c.Request.Context(), invitationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation canceled successfully",
	})
}

// ResendInvitation handles resending an organization invitation
func (h *OrganizationHandler) ResendInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	invitation, err := h.authService.OrganizationService().ResendInvitation(c.Request.Context(), invitationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation resent successfully",
		"data":    invitation,
	})
}

// GetInvitationDetails handles getting invitation details by token
func (h *OrganizationHandler) GetInvitationDetails(c *gin.Context) {
	// For now, return not implemented
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"message": "GetInvitationDetails not yet implemented",
	})
}
