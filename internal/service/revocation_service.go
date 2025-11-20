package service

import (
	"context"
	"fmt"
	"time"

	"auth-service/internal/repository"
	"auth-service/pkg/jwt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RevocationService handles token revocation and JWT denylist
type RevocationService interface {
	// RevokeToken adds a token to the denylist
	RevokeToken(ctx context.Context, tokenString string) error

	// IsTokenRevoked checks if a token is in the denylist
	IsTokenRevoked(ctx context.Context, tokenString string) (bool, error)

	// RevokeUserSessions revokes all active sessions for a user
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error

	// RevokeOrgSessions revokes all active sessions for an organization
	RevokeOrgSessions(ctx context.Context, orgID uuid.UUID) error

	// RevokeUserInOrg revokes all sessions for a user in a specific organization
	RevokeUserInOrg(ctx context.Context, userID, orgID uuid.UUID) error

	// CleanupExpiredTokens removes expired tokens from the denylist
	CleanupExpiredTokens(ctx context.Context) error
}

type revocationService struct {
	repo       repository.Repository
	jwtService jwt.JWTService
	redis      *redis.Client
}

// NewRevocationService creates a new revocation service
func NewRevocationService(repo repository.Repository, jwtService jwt.JWTService, redisClient *redis.Client) RevocationService {
	return &revocationService{
		repo:       repo,
		jwtService: jwtService,
		redis:      redisClient,
	}
}

// RevokeToken adds a token to the Redis denylist
func (s *revocationService) RevokeToken(ctx context.Context, tokenString string) error {
	// Parse token to get expiration and JTI
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Use JTI (JWT ID) as the key for the denylist
	if claims.ID == "" {
		return fmt.Errorf("token has no JTI claim")
	}

	// Calculate TTL based on token expiration
	var ttl time.Duration
	if claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
		if ttl <= 0 {
			// Token already expired, no need to add to denylist
			return nil
		}
	} else {
		// If no expiration, store for 24 hours by default
		ttl = 24 * time.Hour
	}

	// Add to Redis denylist with TTL
	key := fmt.Sprintf("revoked:token:%s", claims.ID)
	err = s.redis.Set(ctx, key, "revoked", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to denylist: %w", err)
	}

	return nil
}

// IsTokenRevoked checks if a token is in the Redis denylist
func (s *revocationService) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	// Parse token to get JTI
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		// Invalid token is considered revoked
		return true, nil
	}

	if claims.ID == "" {
		return false, nil
	}

	// Check if token JTI exists in Redis
	key := fmt.Sprintf("revoked:token:%s", claims.ID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}

	return exists > 0, nil
}

// RevokeUserSessions revokes all active sessions for a user across all organizations
func (s *revocationService) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all active sessions for the user
	sessions, err := s.repo.UserSession().GetByUserID(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("failed to fetch user sessions: %w", err)
	}

	// Revoke each session
	for _, session := range sessions {
		if err := s.repo.UserSession().Delete(ctx, session.ID.String()); err != nil {
			return fmt.Errorf("failed to revoke session %s: %w", session.ID, err)
		}
	}

	// Add user to revocation set in Redis (for 24 hours)
	// This allows middleware to check if user's tokens should be rejected
	key := fmt.Sprintf("revoked:user:%s", userID.String())
	err = s.redis.Set(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to revocation set: %w", err)
	}

	return nil
}

// RevokeOrgSessions revokes all active sessions for all users in an organization
func (s *revocationService) RevokeOrgSessions(ctx context.Context, orgID uuid.UUID) error {
	// Get all memberships for the organization
	memberships, err := s.repo.OrganizationMembership().GetByOrganization(ctx, orgID.String())
	if err != nil {
		return fmt.Errorf("failed to fetch organization memberships: %w", err)
	}

	// Revoke sessions for each user in the organization
	for _, membership := range memberships {
		// Get sessions for this user in this org
		sessions, err := s.repo.UserSession().GetByUserID(ctx, membership.UserID.String())
		if err != nil {
			continue // Skip on error, log in production
		}

		// Filter sessions that are for this organization and revoke them
		for _, session := range sessions {
			if session.OrganizationID == orgID {
				if err := s.repo.UserSession().Delete(ctx, session.ID.String()); err != nil {
					// Log error but continue
					continue
				}
			}
		}
	}

	// Add organization to revocation set in Redis (for 24 hours)
	key := fmt.Sprintf("revoked:org:%s", orgID.String())
	err = s.redis.Set(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to add organization to revocation set: %w", err)
	}

	return nil
}

// RevokeUserInOrg revokes all sessions for a specific user in a specific organization
func (s *revocationService) RevokeUserInOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	// Get all sessions for the user
	sessions, err := s.repo.UserSession().GetByUserID(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("failed to fetch user sessions: %w", err)
	}

	// Revoke only sessions for this user in this organization
	for _, session := range sessions {
		if session.OrganizationID == orgID {
			if err := s.repo.UserSession().Delete(ctx, session.ID.String()); err != nil {
				return fmt.Errorf("failed to revoke session %s: %w", session.ID, err)
			}
		}
	}

	// Add user+org combination to revocation set in Redis (for 24 hours)
	key := fmt.Sprintf("revoked:user_org:%s:%s", userID.String(), orgID.String())
	err = s.redis.Set(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to add user+org to revocation set: %w", err)
	}

	return nil
}

// CleanupExpiredTokens is typically not needed for Redis-based denylist
// since Redis automatically removes keys when they expire (TTL).
// This method can be used for any additional cleanup logic if needed.
func (s *revocationService) CleanupExpiredTokens(ctx context.Context) error {
	// Redis automatically handles TTL expiration
	// This is a no-op but kept for interface completeness
	return nil
}

// IsUserRevoked checks if a user has been globally revoked
func (s *revocationService) IsUserRevoked(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("revoked:user:%s", userID.String())
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user revocation: %w", err)
	}
	return exists > 0, nil
}

// IsOrgRevoked checks if an organization has been globally revoked
func (s *revocationService) IsOrgRevoked(ctx context.Context, orgID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("revoked:org:%s", orgID.String())
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check org revocation: %w", err)
	}
	return exists > 0, nil
}

// IsUserOrgRevoked checks if a user+org combination has been revoked
func (s *revocationService) IsUserOrgRevoked(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("revoked:user_org:%s:%s", userID.String(), orgID.String())
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user+org revocation: %w", err)
	}
	return exists > 0, nil
}
