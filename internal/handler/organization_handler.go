package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
	search := c.Query("search") // Get search parameter

	response, err := h.authService.OrganizationService().ListMembers(c.Request.Context(), orgID, search)
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
	search := c.Query("search") // Get search parameter

	response, err := h.authService.OrganizationService().ListPendingInvitations(c.Request.Context(), orgID, search)
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

// GetInvitationDetails handles getting invitation details by token (public endpoint)
func (h *OrganizationHandler) GetInvitationDetails(c *gin.Context) {
	token := c.Param("token")

	invitation, err := h.authService.OrganizationService().GetInvitationByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Invitation not found or expired",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"email":             invitation.Email,
			"organization_name": invitation.OrganizationName,
			"role_name":         invitation.RoleName,
			"expires_at":        invitation.ExpiresAt,
			"status":            invitation.Status,
		},
	})
}

// GetOrganizationRoles handles getting roles for an organization
// Filtering logic:
// - Superadmin: sees all roles (system + custom)
// - Organization admin: sees ONLY custom org roles (IsSystem=false)
func (h *OrganizationHandler) GetOrganizationRoles(c *gin.Context) {
	orgID := c.Param("orgId")

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid organization ID",
		})
		return
	}

	// Check if user is superadmin
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	// Get all roles for the organization
	allRoles, err := h.authService.RoleService().GetRolesByOrganization(c.Request.Context(), orgUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Filter roles based on user type
	var response interface{}
	if isSuperadmin {
		// Superadmin sees ALL roles (no filtering)
		response = allRoles
	} else {
		// Organization admin: filter out system roles (IsSystem=true)
		// Only show custom org roles (IsSystem=false)
		filteredRoles := make([]*service.RoleResponse, 0)
		for _, role := range allRoles {
			if !role.IsSystem {
				filteredRoles = append(filteredRoles, role)
			}
		}
		response = filteredRoles
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
