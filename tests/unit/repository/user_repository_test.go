package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/tests/testutils"
)

func TestUserRepository_Create(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Email:        "test@example.com",
				PasswordHash: "hashedpassword",
				UserType:     "Admin",
				TenantID:     tenant.ID,
				Status:       models.UserStatusActive,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "duplicate user",
			user: &models.User{
				Email:        "test@example.com", // Same email
				PasswordHash: "hashedpassword2",
				UserType:     "Student",
				TenantID:     tenant.ID,
				Status:       models.UserStatusActive,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: true,
		},
		{
			name:    "nil user",
			user:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean test data
			testutils.CleanTestData(t, testDB.DB)
			tenant = testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")

			err := userRepo.Create(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.user.ID)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hashedpass", "Admin", tenant.ID)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "existing user",
			userID:  user.ID.String(),
			wantErr: false,
		},
		{
			name:    "non-existing user",
			userID:  uuid.New().String(),
			wantErr: true,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedUser, err := userRepo.GetByID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, retrievedUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedUser)
				assert.Equal(t, user.ID, retrievedUser.ID)
				assert.Equal(t, user.Email, retrievedUser.Email)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hashedpass", "Admin", tenant.ID)

	tests := []struct {
		name     string
		email    string
		tenantID string
		wantErr  bool
	}{
		{
			name:     "existing user with domain",
			email:    "test@example.com",
			tenantID: tenant.ID.String(),
			wantErr:  false,
		},
		{
			name:     "non-existing email",
			email:    "nonexistent@example.com",
			tenantID: tenant.ID.String(),
			wantErr:  true,
		},
		{
			name:     "existing email wrong tenant",
			email:    "test@example.com",
			tenantID: uuid.New().String(),
			wantErr:  true,
		},
		{
			name:     "empty email",
			email:    "",
			tenantID: tenant.ID.String(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedUser, err := userRepo.GetByEmail(context.Background(), tt.email, tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, retrievedUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedUser)
				assert.Equal(t, user.Email, retrievedUser.Email)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hashedpass", "Admin", tenant.ID)

	t.Run("successful update", func(t *testing.T) {
		user.Firstname = &[]string{"John"}[0]
		user.Lastname = &[]string{"Doe"}[0]

		err := userRepo.Update(context.Background(), user)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := userRepo.GetByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "John", *updatedUser.Firstname)
		assert.Equal(t, "Doe", *updatedUser.Lastname)
	})

	t.Run("update non-existing user", func(t *testing.T) {
		nonExistingUser := &models.User{
			ID:          uuid.New(),
			Email:       "nonexistent@example.com",
			PasswordHash: "hash",
			UserType:    "Admin",
			TenantID:    tenant.ID,
		}

		err := userRepo.Update(context.Background(), nonExistingUser)
		assert.Error(t, err)
	})
}

func TestUserRepository_List(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")

	// Create multiple users
	testutils.CreateTestUser(t, testDB.DB, "user1@example.com", "hash1", "Admin", tenant.ID)
	testutils.CreateTestUser(t, testDB.DB, "user2@example.com", "hash2", "Student", tenant.ID)
	testutils.CreateTestUser(t, testDB.DB, "user3@example.com", "hash3", "RTO", tenant.ID)

	t.Run("list all users", func(t *testing.T) {
		retrievedUsers, err := userRepo.List(context.Background(), tenant.ID.String(), 10, "")
		assert.NoError(t, err)
		assert.Len(t, retrievedUsers, 3)

		// Check if all users are returned
		emails := make([]string, len(retrievedUsers))
		for i, u := range retrievedUsers {
			emails[i] = u.Email
		}
		assert.Contains(t, emails, "user1@example.com")
		assert.Contains(t, emails, "user2@example.com")
		assert.Contains(t, emails, "user3@example.com")
	})

	t.Run("list with pagination", func(t *testing.T) {
		retrievedUsers, err := userRepo.List(context.Background(), tenant.ID.String(), 2, "")
		assert.NoError(t, err)
		assert.Len(t, retrievedUsers, 2)
	})

	t.Run("list with cursor", func(t *testing.T) {
		// Get first page
		firstPage, err := userRepo.List(context.Background(), tenant.ID.String(), 1, "")
		assert.NoError(t, err)
		assert.Len(t, firstPage, 1)

		// Get second page using cursor from first user
		cursor := firstPage[0].ID.String()
		secondPage, err := userRepo.List(context.Background(), tenant.ID.String(), 2, cursor)
		assert.NoError(t, err)
		assert.Len(t, secondPage, 2)
	})
}

func TestUserRepository_Count(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")

	// Initially should be 0
	count, err := userRepo.Count(context.Background(), tenant.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create users
	testutils.CreateTestUser(t, testDB.DB, "user1@example.com", "hash1", "Admin", tenant.ID)
	testutils.CreateTestUser(t, testDB.DB, "user2@example.com", "hash2", "Student", tenant.ID)

	// Should be 2 now
	count, err = userRepo.Count(context.Background(), tenant.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hashedpass", "Admin", tenant.ID)

	originalLastLogin := user.LastLoginAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond * 10)

	err := userRepo.UpdateLastLogin(context.Background(), user.ID.String())
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := userRepo.GetByID(context.Background(), user.ID.String())
	assert.NoError(t, err)
	assert.NotEqual(t, originalLastLogin, updatedUser.LastLoginAt)
	assert.True(t, updatedUser.LastLoginAt.After(*originalLastLogin))
}

func TestUserRepository_Activate_Deactivate(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hashedpass", "Admin", tenant.ID)

	// Initially active
	retrievedUser, _ := userRepo.GetByID(context.Background(), user.ID.String())
	assert.Equal(t, models.UserStatusActive, retrievedUser.Status)

	// Deactivate
	err := userRepo.Deactivate(context.Background(), user.ID.String())
	assert.NoError(t, err)

	retrievedUser, _ = userRepo.GetByID(context.Background(), user.ID.String())
	assert.Equal(t, models.UserStatusDeactivated, retrievedUser.Status)

	// Activate again
	err = userRepo.Activate(context.Background(), user.ID.String())
	assert.NoError(t, err)

	retrievedUser, _ = userRepo.GetByID(context.Background(), user.ID.String())
	assert.Equal(t, models.UserStatusActive, retrievedUser.Status)
}