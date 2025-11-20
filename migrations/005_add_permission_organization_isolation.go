//go:build ignore
// Migration 005: Permission Organization Isolation
// This migration adds organization_id to permissions table to enable custom permission isolation
//
// HANDLED BY: GORM AutoMigrate in internal/repository/repository.go Migrate()
//
// Schema Changes:
// ---------------
// MODIFIED TABLES:
// - permissions: ADDED organization_id UUID (nullable, references organizations.id ON DELETE CASCADE)
//   * System permissions: organization_id = NULL (accessible to all organizations)
//   * Custom permissions: organization_id = specific org UUID (accessible only to that organization)
//
// INDEXES ADDED:
// - idx_permissions_organization_id: Index on organization_id for performance
// - idx_permissions_name_organization: Composite index on (name, organization_id) for lookups
//
// CONSTRAINTS ADDED:
// - permissions_name_organization_unique: UNIQUE(name, organization_id)
//   * Prevents duplicate permission names within same organization context
//   * Allows same permission name as system permission (NULL) and in different organizations
//
// Security Benefits:
// ------------------
// 1. Custom permissions are isolated per organization - cannot be used cross-organization
// 2. System permissions remain globally accessible to all organizations
// 3. Permission assignment now validates organization context before allowing access
// 4. Repository methods use organization-scoped queries for custom permission operations
//
// Service Layer Changes:
// ----------------------
// - Added organization-scoped permission methods: GetByNamesAndOrganization, ListAllForOrganization
// - Role permission assignment now validates permission accessibility within organization
// - Custom permission creation tied to specific organization via CreatePermissionForOrganization
//
// Original Issue Fixed:
// ---------------------
// "hindi dapat magagamit ang custom custom role and permission sa ibang organization"
// Users can no longer see or assign custom roles/permissions from other organizations

package main

// This file documents the permission organization isolation migration.
// The actual migration is performed by GORM AutoMigrate when Permission model changes are detected.
