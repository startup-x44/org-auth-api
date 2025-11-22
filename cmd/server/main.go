package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/seeder"
	"auth-service/internal/service"
	"auth-service/pkg/email"
	"auth-service/pkg/hashutil"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
	"auth-service/pkg/metrics"
	"auth-service/pkg/password"
	"auth-service/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	gormtracing "gorm.io/plugin/opentelemetry/tracing"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logging
	loggerCfg := &logger.Config{
		Level:        cfg.Logging.Level,
		Format:       cfg.Logging.Format,
		Output:       os.Stdout,
		EnableCaller: true,
	}
	logger.Initialize(loggerCfg)
	logger.InfoMsg("Starting auth-service", map[string]interface{}{
		"environment": cfg.Environment,
		"port":        cfg.Server.Port,
	})

	// Initialize HMAC secret for deterministic token hashing
	if err := hashutil.InitializeHMACSecret(); err != nil {
		logger.FatalMsg("Failed to initialize HMAC secret", err)
	}

	// Initialize database
	db := initDatabase(cfg)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	logger.InfoMsg("Database connected successfully")

	// Initialize Redis
	redisClient := initRedis(cfg)
	defer redisClient.Close()

	logger.InfoMsg("Redis connected successfully")

	// Initialize distributed tracing
	tracerProvider, err := tracing.Initialize(&tracing.Config{
		Enabled:      cfg.Tracing.Enabled,
		ServiceName:  cfg.Tracing.ServiceName,
		Environment:  cfg.Environment,
		ExporterType: cfg.Tracing.ExporterType,
		OTLPEndpoint: cfg.Tracing.OTLPEndpoint,
		OTLPInsecure: cfg.Tracing.OTLPInsecure,
		SamplingRate: cfg.Tracing.SamplingRate,
	})
	if err != nil {
		logger.FatalMsg("Failed to initialize tracing", err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.ErrorMsg("Failed to shutdown tracer provider", err)
		}
	}()
	if cfg.Tracing.Enabled {
		logger.InfoMsg("Distributed tracing initialized", map[string]interface{}{
			"exporter":      cfg.Tracing.ExporterType,
			"sampling_rate": cfg.Tracing.SamplingRate,
		})
	}

	// Initialize Prometheus metrics
	metrics.Initialize()
	logger.InfoMsg("Prometheus metrics initialized")

	// Start metrics collector for periodic gauge updates
	metricsCollector := metrics.NewCollector(db, redisClient, 30*time.Second)
	go metricsCollector.Start(context.Background())
	defer metricsCollector.Stop()

	// Initialize repositories
	repo := repository.NewRepository(db)

	// Initialize services
	jwtService, err := jwt.NewService(&cfg.JWT)
	if err != nil {
		logger.FatalMsg("Failed to initialize JWT service", err)
	}
	passwordService := password.NewService()
	emailSvc := email.NewService(&cfg.Email)
	authService := service.NewAuthService(repo, jwtService, passwordService, emailSvc, redisClient)

	// Set Redis and Email clients for user service
	userSvc := authService.UserService()
	userSvc.SetRedisClient(redisClient)
	userSvc.SetEmailService(emailSvc)

	// Initialize OAuth2 services
	clientAppService := service.NewClientAppService(repo)
	oauth2Service := service.NewOAuth2Service(repo, jwtService)
	apiKeyService := service.NewAPIKeyService(repo.APIKey())

	// Initialize audit service
	auditService := service.NewAuditService(db)

	// Get database/SQL connection for health checks
	sqlDB, err := db.DB()
	if err != nil {
		logger.FatalMsg("Failed to get SQL DB instance", err)
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, auditService)
	adminHandler := handler.NewAdminHandler(authService)
	organizationHandler := handler.NewOrganizationHandler(authService)
	roleHandler := handler.NewRoleHandler(authService, auditService)
	rbacHandler := handler.NewRBACHandler(authService.RoleService())
	clientAppHandler := handler.NewClientAppHandler(clientAppService)
	oauth2Handler := handler.NewOAuth2Handler(oauth2Service, clientAppService, userSvc)
	oauthAuditHandler := handler.NewOAuthAuditHandler(db)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)
	revocationHandler := handler.NewRevocationHandler(authService.RevocationService())
	healthHandler := handler.NewHealthHandler(sqlDB, redisClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, repo)
	organizationMiddleware := middleware.NewOrganizationMiddleware(authService)
	rateLimiter := middleware.NewRateLimiter(redisClient, &cfg.RateLimit)
	revocationMiddleware := middleware.RevocationMiddleware(jwtService, authService.RevocationService())

	// Initialize Gin router
	router := setupRouter(cfg, authHandler, adminHandler, organizationHandler, roleHandler, rbacHandler, clientAppHandler, oauth2Handler, oauthAuditHandler, apiKeyHandler, revocationHandler, healthHandler, authMiddleware, organizationMiddleware, rateLimiter, revocationMiddleware)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.InfoMsg("Server started successfully", map[string]interface{}{
			"port": cfg.Server.Port,
			"host": cfg.Server.Host,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalMsg("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.InfoMsg("Received shutdown signal, shutting down gracefully...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.FatalMsg("Server forced to shutdown", err)
	}

	logger.InfoMsg("Server exited cleanly")
}

func initDatabase(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.FatalMsg("Failed to connect to database", err)
	}

	// Add OpenTelemetry tracing to GORM
	if err := db.Use(gormtracing.NewPlugin()); err != nil {
		logger.FatalMsg("Failed to add GORM tracing plugin", err)
	}

	// Auto migrate the schema
	if err := repository.Migrate(db); err != nil {
		logger.FatalMsg("Failed to migrate database", err)
	}

	// Run database seeders
	if err := runSeeders(db); err != nil {
		logger.FatalMsg("Failed to run database seeders", err)
	}

	return db
}

func initRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Add OpenTelemetry tracing to Redis BEFORE any commands (including Ping)
	rdb.AddHook(redisotel.NewTracingHook())

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.FatalMsg("Failed to connect to Redis", err)
	}

	return rdb
}

func runSeeders(db *gorm.DB) error {
	seeder := seeder.NewDatabaseSeeder(db)
	ctx := context.Background()
	return seeder.Seed(ctx)
}

func setupRouter(cfg *config.Config, authHandler *handler.AuthHandler, adminHandler *handler.AdminHandler, organizationHandler *handler.OrganizationHandler, roleHandler *handler.RoleHandler, rbacHandler *handler.RBACHandler, clientAppHandler *handler.ClientAppHandler, oauth2Handler *handler.OAuth2Handler, oauthAuditHandler *handler.OAuthAuditHandler, apiKeyHandler *handler.APIKeyHandler, revocationHandler *handler.RevocationHandler, healthHandler *handler.HealthHandler, authMiddleware *middleware.AuthMiddleware, organizationMiddleware *middleware.OrganizationMiddleware, rateLimiter *middleware.RateLimiter, revocationMiddleware gin.HandlerFunc) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware - CRITICAL ORDER:
	// 1. Recovery MUST be first to catch all panics
	// 2. Tracing MUST be second to wrap everything
	// 3. Security headers before logging/metrics
	// 4. Then logging, metrics, CORS
	// 5. Revocation check early
	// 6. CSRF last (after auth)

	router.Use(gin.Recovery()) // Panic recovery (MUST be first)
	router.Use(middleware.TracingMiddleware(cfg.Tracing.ServiceName, middleware.TracingConfig{
		ExcludePaths: []string{
			"/health",
			"/health/live",
			"/health/ready",
			"/metrics",
		},
	})) // Distributed tracing (MUST wrap everything, exclude health/metrics)

	// Add HTTP status + error recording and X-Trace-ID header
	router.Use(func(c *gin.Context) {
		c.Next() // Process request first

		span := trace.SpanFromContext(c.Request.Context())
		if !span.IsRecording() {
			return
		}

		// Add X-Trace-ID to response headers AFTER span is finalized
		spanCtx := span.SpanContext()
		if spanCtx.IsValid() {
			c.Header("X-Trace-ID", spanCtx.TraceID().String())
		}

		// Record HTTP status
		status := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", status))

		// Only record 5xx errors, not 4xx (client errors are not trace errors)
		if status >= 500 {
			if len(c.Errors) > 0 {
				// Record each error individually
				for _, ginErr := range c.Errors {
					span.RecordError(ginErr.Err)
				}
				span.SetStatus(codes.Error, c.Errors.String())
			} else {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
			}
		}
	})

	router.Use(middleware.SecurityHeadersMiddleware())             // Security headers MUST be early
	router.Use(middleware.StructuredLoggingMiddleware())           // Structured logging with request_id
	router.Use(middleware.MetricsMiddleware())                     // Prometheus metrics collection
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins)) // CORS
	router.Use(revocationMiddleware)                               // Check token revocation early

	// CSRF middleware for state-changing operations
	csrfConfig := middleware.DefaultCSRFConfig(cfg.JWT.Secret, cfg.Environment == "production")
	csrfConfig.SkipPaths = []string{
		"/api/v1/oauth/token",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
	}
	router.Use(middleware.CSRFMiddleware(csrfConfig))

	// Health check endpoints (Kubernetes-compatible)
	router.GET("/health", healthHandler.HealthCheck)          // Legacy comprehensive health check
	router.GET("/health/live", healthHandler.LivenessProbe)   // Kubernetes liveness probe
	router.GET("/health/ready", healthHandler.ReadinessProbe) // Kubernetes readiness probe

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes with rate limiting
		auth := v1.Group("/auth")
		{
			auth.POST("/register", rateLimiter.ByIP(middleware.ScopeRegistration), authHandler.RegisterGlobal)
			auth.POST("/login", rateLimiter.ByIP(middleware.ScopeLogin), authHandler.LoginGlobal)
			auth.POST("/refresh", rateLimiter.ByUserID(middleware.ScopeTokenRefresh), authHandler.RefreshToken)
			auth.POST("/forgot-password", rateLimiter.ByEmail(middleware.ScopePasswordReset, "email"), authHandler.ForgotPassword)
			auth.POST("/reset-password", rateLimiter.ByIP(middleware.ScopePasswordReset), authHandler.ResetPassword)
			auth.POST("/verify-email", rateLimiter.ByIP(middleware.ScopeRegistration), authHandler.VerifyEmail)
			auth.POST("/resend-verification", rateLimiter.ByEmail(middleware.ScopePasswordReset, "email"), authHandler.ResendVerificationEmail)
		}

		// Organization selection (requires valid credentials from login)
		auth.POST("/select-organization", authHandler.SelectOrganization)
		auth.POST("/create-organization", authHandler.CreateOrganization)

		// Protected user routes
		user := v1.Group("/user")
		user.Use(authMiddleware.AuthRequired())
		user.Use(middleware.AddUserContextMiddleware())          // Add user context to logs
		user.Use(rateLimiter.ByUserID(middleware.ScopeAPICalls)) // General API rate limiting
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
		org.Use(middleware.AddUserContextMiddleware()) // Add user context to logs
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
			org.POST("/:orgId/roles", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("role:create"), roleHandler.CreateRole)
			org.GET("/:orgId/roles/:roleId", organizationMiddleware.MembershipRequired(""), roleHandler.GetRole)
			org.PUT("/:orgId/roles/:roleId", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("role:update"), roleHandler.UpdateRole)
			org.DELETE("/:orgId/roles/:roleId", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("role:delete"), roleHandler.DeleteRole)

			// Role permissions
			org.POST("/:orgId/roles/:roleId/permissions", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("role:update"), roleHandler.AssignPermissions)
			org.DELETE("/:orgId/roles/:roleId/permissions", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("role:update"), roleHandler.RevokePermissions)

			// List all available permissions
			org.GET("/:orgId/permissions", organizationMiddleware.MembershipRequired(""), roleHandler.ListPermissions)
			org.POST("/:orgId/permissions", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("permission:create"), roleHandler.CreatePermission)
			org.PUT("/:orgId/permissions/:permission_id", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("permission:update"), roleHandler.UpdatePermission)
			org.DELETE("/:orgId/permissions/:permission_id", organizationMiddleware.MembershipRequired(""), authMiddleware.RequirePermission("permission:delete"), roleHandler.DeletePermission)

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
		admin.Use(authMiddleware.LoadUser())
		admin.Use(authMiddleware.AdminRequired())
		{
			// Global user management
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:userId/activate", adminHandler.ActivateUser)
			admin.PUT("/users/:userId/deactivate", adminHandler.DeactivateUser)
			admin.DELETE("/users/:userId", adminHandler.DeleteUser)

			// Global organization management
			admin.GET("/organizations", adminHandler.ListOrganizations)

			// RBAC management (superadmin only)
			rbac := admin.Group("/rbac")
			{
				// Permissions
				rbac.GET("/permissions", rbacHandler.ListAllPermissions)

				// System roles
				rbac.GET("/roles", rbacHandler.ListSystemRoles)
				rbac.POST("/roles", rbacHandler.CreateSystemRole)
				rbac.GET("/roles/:id", rbacHandler.GetSystemRole)
				rbac.PUT("/roles/:id", rbacHandler.UpdateSystemRole)
				rbac.DELETE("/roles/:id", rbacHandler.DeleteSystemRole)

				// Role permissions
				rbac.GET("/roles/:id/permissions", rbacHandler.GetRolePermissions)
				rbac.POST("/roles/:id/permissions", rbacHandler.AssignPermissionsToRole)
				rbac.DELETE("/roles/:id/permissions", rbacHandler.RevokePermissionsFromRole)

				// Statistics
				rbac.GET("/stats", rbacHandler.GetRBACStats)
			}

			// Client application management (superadmin only)
			admin.POST("/client-apps", clientAppHandler.CreateClientApp)
			admin.GET("/client-apps", clientAppHandler.ListClientApps)
			admin.GET("/client-apps/:id", clientAppHandler.GetClientApp)
			admin.PUT("/client-apps/:id", clientAppHandler.UpdateClientApp)
			admin.DELETE("/client-apps/:id", clientAppHandler.DeleteClientApp)
			admin.POST("/client-apps/:id/rotate-secret", clientAppHandler.RotateClientSecret)
		}

		// OAuth2 endpoints (public/authenticated as needed)
		oauth := v1.Group("/oauth")
		{
			// Authorization endpoint (requires authentication)
			oauth.GET("/authorize", authMiddleware.AuthRequired(), oauth2Handler.Authorize)

			// Token endpoint (public - validates client credentials) with rate limiting
			oauth.POST("/token", rateLimiter.ByIP(middleware.ScopeOAuth2Token), oauth2Handler.Token)

			// UserInfo endpoint (requires valid OAuth2 access token)
			oauth.GET("/userinfo", authMiddleware.AuthRequired(), oauth2Handler.UserInfo)

			// Logout endpoint (requires authentication)
			oauth.POST("/logout", authMiddleware.AuthRequired(), oauth2Handler.Logout)

			// Audit endpoints (superadmin only)
			audit := oauth.Group("/audit")
			audit.Use(authMiddleware.AuthRequired())
			audit.Use(authMiddleware.LoadUser())
			audit.Use(authMiddleware.AdminRequired())
			{
				audit.GET("/authorizations", oauthAuditHandler.ListAuthorizationLogs)
				audit.GET("/tokens", oauthAuditHandler.ListTokenGrants)
				audit.GET("/stats", oauthAuditHandler.GetAuditStats)
			}
		}

		// Developer API endpoints (protected routes)
		dev := v1.Group("/dev")
		dev.Use(authMiddleware.AuthRequired())
		{
			handler.RegisterAPIKeyRoutes(dev, apiKeyHandler)
		}

		// Token revocation endpoints (requires authentication)
		revocation := v1.Group("/revocation")
		revocation.Use(authMiddleware.AuthRequired())
		{
			// Revoke a specific token
			revocation.POST("/token", revocationHandler.RevokeToken)

			// Revoke all sessions for a user (requires admin or self)
			revocation.DELETE("/user/:user_id/sessions", authMiddleware.AdminRequired(), revocationHandler.RevokeUserSessions)

			// Revoke all sessions for an organization (requires org admin)
			revocation.DELETE("/org/:org_id/sessions", organizationMiddleware.OrgAdminRequired(), revocationHandler.RevokeOrgSessions)

			// Revoke user sessions in specific org (requires org admin)
			revocation.DELETE("/user/:user_id/org/:org_id/sessions", organizationMiddleware.OrgAdminRequired(), revocationHandler.RevokeUserInOrg)
		}
	}

	return router
}
