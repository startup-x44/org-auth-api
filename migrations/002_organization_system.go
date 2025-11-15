// Migration 002: Organization System
// This migration transforms the auth service from tenant-based to Slack-style multi-organization architecture
//
// HANDLED BY: GORM AutoMigrate in internal/repository/repository.go Migrate()
//
// Schema Changes:
// ---------------
// ADDED TABLES:
// - organizations: Core organization/workspace table (id, name, slug, settings, created_at, etc.)
// - organization_memberships: User-Org relationships (user_id, organization_id, role, status)
//   * Unique constraint: (organization_id, user_id)
// - organization_invitations: Pending invitations (organization_id, email, token_hash, role, expires_at)
//   * Unique constraint: (organization_id, email)
//
// MODIFIED TABLES:
// - users: REMOVED tenant_id (users are now global)
// - user_sessions: tenant_id → organization_id (sessions are org-scoped)
// - refresh_tokens: tenant_id → organization_id (tokens are org-scoped)
// - password_resets: Remains global (no tenant/org field)
// - failed_login_attempts: Remains global (email + IP based)
//
// REMOVED TABLES:
// - tenants (entire tenant model removed)
//
// Auth Flow Changes:
// ------------------
// 1. User registers globally (no org required)
// 2. User logs in → receives list of their organizations
// 3. User selects organization → receives org-scoped JWT
// 4. User can create new organizations or accept invitations
// 5. JWT contains OrganizationID and OrganizationRole claims

package main

// This file documents the organization system migration.
// The actual migration is performed by GORM AutoMigrate.
