package seeder

import (
	"context"
)

// seedOrganizations creates test organizations with their custom owner roles
// NOTE: Skipped for superadmin-only setup. Organizations are created through user registration.
func (s *DatabaseSeeder) seedOrganizations(ctx context.Context) error {
	// Organizations should be created through the normal registration flow
	// where users create their own organizations after registering.
	// This ensures proper multi-tenant isolation and role assignment.

	// For superadmin testing, organizations are not needed since superadmin
	// bypasses the organization selection flow and has system-wide access.

	return nil
}
