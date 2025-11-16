package handler

import (
	"net/http"
	"strconv"
	"time"

	"auth-service/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OAuthAuditHandler handles OAuth audit log endpoints
type OAuthAuditHandler struct {
	db *gorm.DB
}

// NewOAuthAuditHandler creates a new OAuth audit handler
func NewOAuthAuditHandler(db *gorm.DB) *OAuthAuditHandler {
	return &OAuthAuditHandler{
		db: db,
	}
}

// AuthorizationLogEntry represents an authorization attempt
type AuthorizationLogEntry struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	ClientName     string    `json:"client_name"`
	UserID         string    `json:"user_id"`
	UserEmail      string    `json:"user_email"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Scope          string    `json:"scope"`
	Status         string    `json:"status"` // "granted", "denied", "expired", "used"
	CreatedAt      time.Time `json:"created_at"`
}

// TokenLogEntry represents a token grant
type TokenLogEntry struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	ClientName     string    `json:"client_name"`
	UserID         string    `json:"user_id"`
	UserEmail      string    `json:"user_email"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Scope          string    `json:"scope"`
	GrantType      string    `json:"grant_type"` // "authorization_code", "refresh_token"
	IsActive       bool      `json:"is_active"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListAuthorizationLogs godoc
// @Summary List OAuth authorization attempts
// @Tags oauth-audit
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Param client_id query string false "Filter by client ID"
// @Param user_id query string false "Filter by user ID"
// @Success 200 {object} gin.H{data=[]AuthorizationLogEntry,total=int}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /oauth/audit/authorizations [get]
func (h *OAuthAuditHandler) ListAuthorizationLogs(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user := userVal.(*models.User)
	if !user.IsSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "superadmin access required",
		})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	clientID := c.Query("client_id")
	userID := c.Query("user_id")

	// Build query
	query := h.db.Table("authorization_codes").
		Select(`
			authorization_codes.id,
			authorization_codes.client_id,
			client_apps.name as client_name,
			authorization_codes.user_id,
			users.email as user_email,
			authorization_codes.organization_id,
			authorization_codes.scope,
			CASE 
				WHEN authorization_codes.used THEN 'used'
				WHEN authorization_codes.expires_at < NOW() THEN 'expired'
				ELSE 'active'
			END as status,
			authorization_codes.created_at
		`).
		Joins("LEFT JOIN client_apps ON client_apps.client_id = authorization_codes.client_id").
		Joins("LEFT JOIN users ON users.id = authorization_codes.user_id").
		Order("authorization_codes.created_at DESC")

	// Apply filters
	if clientID != "" {
		query = query.Where("authorization_codes.client_id = ?", clientID)
	}
	if userID != "" {
		query = query.Where("authorization_codes.user_id = ?", userID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var logs []AuthorizationLogEntry
	if err := query.Limit(limit).Offset(offset).Scan(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to fetch authorization logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// ListTokenGrants godoc
// @Summary List OAuth token grants (refresh tokens)
// @Tags oauth-audit
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Param client_id query string false "Filter by client ID"
// @Param user_id query string false "Filter by user ID"
// @Success 200 {object} gin.H{data=[]TokenLogEntry,total=int}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /oauth/audit/tokens [get]
func (h *OAuthAuditHandler) ListTokenGrants(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user := userVal.(*models.User)
	if !user.IsSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "superadmin access required",
		})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	clientID := c.Query("client_id")
	userID := c.Query("user_id")

	// Build query
	query := h.db.Table("oauth_refresh_tokens").
		Select(`
			oauth_refresh_tokens.id,
			oauth_refresh_tokens.client_id,
			client_apps.name as client_name,
			oauth_refresh_tokens.user_id,
			users.email as user_email,
			oauth_refresh_tokens.organization_id,
			oauth_refresh_tokens.scope,
			'refresh_token' as grant_type,
			NOT oauth_refresh_tokens.revoked AND oauth_refresh_tokens.expires_at > NOW() as is_active,
			oauth_refresh_tokens.expires_at,
			oauth_refresh_tokens.created_at
		`).
		Joins("LEFT JOIN client_apps ON client_apps.client_id = oauth_refresh_tokens.client_id").
		Joins("LEFT JOIN users ON users.id = oauth_refresh_tokens.user_id").
		Order("oauth_refresh_tokens.created_at DESC")

	// Apply filters
	if clientID != "" {
		query = query.Where("oauth_refresh_tokens.client_id = ?", clientID)
	}
	if userID != "" {
		query = query.Where("oauth_refresh_tokens.user_id = ?", userID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var logs []TokenLogEntry
	if err := query.Limit(limit).Offset(offset).Scan(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to fetch token grants",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetAuditStats godoc
// @Summary Get OAuth audit statistics
// @Tags oauth-audit
// @Produce json
// @Success 200 {object} gin.H{data=map[string]interface{}}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /oauth/audit/stats [get]
func (h *OAuthAuditHandler) GetAuditStats(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user := userVal.(*models.User)
	if !user.IsSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "superadmin access required",
		})
		return
	}

	stats := make(map[string]interface{})

	// Total authorizations
	var totalAuthorizations int64
	h.db.Model(&models.AuthorizationCode{}).Count(&totalAuthorizations)
	stats["total_authorizations"] = totalAuthorizations

	// Authorizations today
	var authorizationsToday int64
	h.db.Model(&models.AuthorizationCode{}).
		Where("created_at >= ?", time.Now().Truncate(24*time.Hour)).
		Count(&authorizationsToday)
	stats["authorizations_today"] = authorizationsToday

	// Active refresh tokens
	var activeTokens int64
	h.db.Model(&models.OAuthRefreshToken{}).
		Where("revoked = ? AND expires_at > ?", false, time.Now()).
		Count(&activeTokens)
	stats["active_tokens"] = activeTokens

	// Total refresh tokens
	var totalTokens int64
	h.db.Model(&models.OAuthRefreshToken{}).Count(&totalTokens)
	stats["total_tokens"] = totalTokens

	// Unique users
	var uniqueUsers int64
	h.db.Model(&models.AuthorizationCode{}).
		Distinct("user_id").
		Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	// Unique client apps
	var uniqueClients int64
	h.db.Model(&models.AuthorizationCode{}).
		Distinct("client_id").
		Count(&uniqueClients)
	stats["unique_clients"] = uniqueClients

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
