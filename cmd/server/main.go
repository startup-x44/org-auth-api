package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/seeder"
	"auth-service/internal/service"
	"auth-service/pkg/email"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db := initDatabase(cfg)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Initialize Redis
	redisClient := initRedis(cfg)
	defer redisClient.Close()

	// Initialize repositories
	repo := repository.NewRepository(db)

	// Initialize services
	jwtService, err := jwt.NewService(&cfg.JWT)
	if err != nil {
		log.Fatal("Failed to initialize JWT service:", err)
	}
	passwordService := password.NewService()
	emailSvc := email.NewService(&cfg.Email)
	authService := service.NewAuthService(repo, jwtService, passwordService, emailSvc)

	// Set Redis and Email clients for user service
	userSvc := authService.UserService()
	userSvc.SetRedisClient(redisClient)
	userSvc.SetEmailService(emailSvc)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(authService)
	organizationHandler := handler.NewOrganizationHandler(authService)
	roleHandler := handler.NewRoleHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	organizationMiddleware := middleware.NewOrganizationMiddleware(authService)

	// Initialize Gin router
	router := setupRouter(cfg, authHandler, adminHandler, organizationHandler, roleHandler, authMiddleware, organizationMiddleware)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func initDatabase(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	if err := repository.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Run database seeders
	if err := runSeeders(db); err != nil {
		log.Fatalf("Failed to run database seeders: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
}

func initRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return rdb
}

func runSeeders(db *gorm.DB) error {
	seeder := seeder.NewDatabaseSeeder(db)
	ctx := context.Background()
	return seeder.Seed(ctx)
}

func setupRouter(cfg *config.Config, authHandler *handler.AuthHandler, adminHandler *handler.AdminHandler, organizationHandler *handler.OrganizationHandler, roleHandler *handler.RoleHandler, authMiddleware *middleware.AuthMiddleware, organizationMiddleware *middleware.OrganizationMiddleware) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// CSRF middleware for state-changing operations
	csrfConfig := middleware.DefaultCSRFConfig(cfg.JWT.Secret, cfg.Environment == "production")
	router.Use(middleware.CSRFMiddleware(csrfConfig))

	// Health check
	router.GET("/health", authHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.RegisterGlobal)
			auth.POST("/login", authHandler.LoginGlobal)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerificationEmail)
		}

		// Organization selection (requires valid credentials from login)
		auth.POST("/select-organization", authHandler.SelectOrganization)
		auth.POST("/create-organization", authHandler.CreateOrganization)

		// Protected user routes
		user := v1.Group("/user")
		user.Use(authMiddleware.AuthRequired())
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.POST("/change-password", authHandler.ChangePassword)
			user.POST("/logout", authHandler.Logout)
			user.GET("/organizations", authHandler.GetMyOrganizations)
		}

		// Organization routes
		org := v1.Group("/organizations")
		org.Use(authMiddleware.AuthRequired())
		{
			org.POST("", organizationHandler.CreateOrganization)
			org.GET("", organizationHandler.ListUserOrganizations)
			org.GET("/:orgId", organizationMiddleware.MembershipRequired(""), organizationHandler.GetOrganization)

			// Organization management (admin only)
			org.PUT("/:orgId", organizationMiddleware.OrgAdminRequired(), organizationHandler.UpdateOrganization)
			org.DELETE("/:orgId", organizationMiddleware.OrgAdminRequired(), organizationHandler.DeleteOrganization)

			// Organization members
			org.GET("/:orgId/members", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("member:view"), organizationHandler.ListOrganizationMembers)
			org.POST("/:orgId/members", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("member:invite"), organizationHandler.InviteUser)
			org.PUT("/:orgId/members/:userId", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("member:update"), organizationHandler.UpdateMembership)
			org.DELETE("/:orgId/members/:userId", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("member:update"), organizationHandler.RemoveMember)

			// Organization roles
			org.GET("/:orgId/roles", organizationMiddleware.MembershipRequired(""), organizationHandler.GetOrganizationRoles)
			org.POST("/:orgId/roles", organizationMiddleware.OrgAdminRequired(), roleHandler.CreateRole)
			org.GET("/:orgId/roles/:roleId", organizationMiddleware.MembershipRequired(""), roleHandler.GetRole)
			org.PUT("/:orgId/roles/:roleId", organizationMiddleware.OrgAdminRequired(), roleHandler.UpdateRole)
			org.DELETE("/:orgId/roles/:roleId", organizationMiddleware.OrgAdminRequired(), roleHandler.DeleteRole)

			// Role permissions
			org.POST("/:orgId/roles/:roleId/permissions", organizationMiddleware.OrgAdminRequired(), roleHandler.AssignPermissions)
			org.DELETE("/:orgId/roles/:roleId/permissions", organizationMiddleware.OrgAdminRequired(), roleHandler.RevokePermissions)

			// List all available permissions
			org.GET("/:orgId/permissions", organizationMiddleware.MembershipRequired(""), roleHandler.ListPermissions)
			org.POST("/:orgId/permissions", organizationMiddleware.OrgAdminRequired(), roleHandler.CreatePermission)
			org.PUT("/:orgId/permissions/:permission_id", organizationMiddleware.OrgAdminRequired(), roleHandler.UpdatePermission)
			org.DELETE("/:orgId/permissions/:permission_id", organizationMiddleware.OrgAdminRequired(), roleHandler.DeletePermission)

			// Organization invitations
			org.GET("/:orgId/invitations", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("invitation:view"), organizationHandler.GetOrganizationInvitations)
			org.POST("/:orgId/invitations/:invitationId/resend", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("invitation:resend"), organizationHandler.ResendInvitation)
			org.DELETE("/:orgId/invitations/:invitationId", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("invitation:cancel"), organizationHandler.CancelInvitation)
		}

		// Invitation acceptance (requires authentication but NOT organization membership)
		invitations := v1.Group("/invitations")
		invitations.Use(authMiddleware.AuthRequired())
		{
			invitations.POST("/:token/accept", organizationHandler.AcceptInvitation)
		}

		// Public invitation details (no auth required)
		v1.GET("/invitations/:token", organizationHandler.GetInvitationDetails)

		// Protected admin routes (superadmin only)
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.AuthRequired())
		admin.Use(authMiddleware.AdminRequired())
		{
			// Global user management
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:userId/activate", adminHandler.ActivateUser)
			admin.PUT("/users/:userId/deactivate", adminHandler.DeactivateUser)
			admin.DELETE("/users/:userId", adminHandler.DeleteUser)
		}
	}

	return router
}
