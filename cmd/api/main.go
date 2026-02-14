package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/handler"
	"github.com/adverax/crm/internal/middleware"
	"github.com/adverax/crm/internal/modules/auth"
	"github.com/adverax/crm/internal/pkg/config"
	"github.com/adverax/crm/internal/pkg/database"
	"github.com/adverax/crm/internal/platform/dml"
	dmlengine "github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/metadata/ddl"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/fls"
	"github.com/adverax/crm/internal/platform/security/ols"
	"github.com/adverax/crm/internal/platform/security/rls"
	"github.com/adverax/crm/internal/platform/soql"
	soqlengine "github.com/adverax/crm/internal/platform/soql/engine"
	"github.com/adverax/crm/internal/platform/templates"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(config.Load().LogLevel),
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DB.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected", "host", cfg.DB.Host, "db", cfg.DB.Name)

	router := setupRouter(pool, cfg)

	// Start outbox worker
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	startOutboxWorker(workerCtx, pool, cfg.DB.DSN(), logger)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func setupRouter(pool *pgxpool.Pool, cfg config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// --- Metadata layer ---
	objectRepo := metadata.NewPgObjectRepository(pool)
	fieldRepo := metadata.NewPgFieldRepository(pool)
	polyRepo := metadata.NewPgPolymorphicTargetRepository(pool)
	ddlExec := ddl.NewExecutor()

	cacheLoader := metadata.NewPgCacheLoader(pool)
	metadataCache := metadata.NewMetadataCache(cacheLoader)

	ctx := context.Background()
	if err := metadataCache.Load(ctx); err != nil {
		slog.Warn("metadata cache initial load failed (empty database?)", "error", err)
	}

	objectService := metadata.NewObjectService(pool, objectRepo, fieldRepo, ddlExec, metadataCache)
	fieldService := metadata.NewFieldService(pool, objectRepo, fieldRepo, polyRepo, ddlExec, metadataCache)

	metadataHandler := handler.NewMetadataHandler(objectService, fieldService)

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- Security layer ---
	userRoleRepo := security.NewPgUserRoleRepository(pool)
	psRepo := security.NewPgPermissionSetRepository(pool)
	profileRepo := security.NewPgProfileRepository(pool)
	userRepo := security.NewPgUserRepository(pool)
	psToUserRepo := security.NewPgPermissionSetToUserRepository(pool)
	objPermRepo := security.NewPgObjectPermissionRepository(pool)
	fieldPermRepo := security.NewPgFieldPermissionRepository(pool)
	effectiveRepo := security.NewPgEffectivePermissionRepository(pool)
	outboxRepo := security.NewPgOutboxRepository(pool)
	groupRepo := security.NewPgGroupRepository(pool)
	memberRepo := security.NewPgGroupMemberRepository(pool)

	sharingRuleRepo := security.NewPgSharingRuleRepository(pool)

	roleService := security.NewUserRoleService(pool, userRoleRepo, groupRepo, outboxRepo)
	psService := security.NewPermissionSetService(pool, psRepo)
	profileService := security.NewProfileService(pool, profileRepo, psRepo)
	userService := security.NewUserService(pool, userRepo, profileRepo, userRoleRepo, psToUserRepo, outboxRepo, groupRepo, memberRepo)
	permissionService := security.NewPermissionService(pool, objPermRepo, fieldPermRepo, outboxRepo)
	groupService := security.NewGroupService(pool, groupRepo, memberRepo, outboxRepo)
	sharingRuleService := security.NewSharingRuleService(pool, sharingRuleRepo, groupRepo, outboxRepo)

	// --- Auth module ---
	userAuthRepo := auth.NewPgUserAuthRepository(pool)
	refreshTokenRepo := auth.NewPgRefreshTokenRepository(pool)
	resetTokenRepo := auth.NewPgPasswordResetTokenRepository(pool)
	emailSender := auth.NewConsoleEmailSender()
	rateLimiter := auth.NewRateLimiter(5, 15*time.Minute)

	jwtSecret := cfg.JWT.Secret
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	authService := auth.NewService(auth.ServiceConfig{
		UserRepo:     userAuthRepo,
		RefreshRepo:  refreshTokenRepo,
		ResetRepo:    resetTokenRepo,
		EmailSender:  emailSender,
		JWTSecret:    jwtSecret,
		AccessTTL:    cfg.JWT.AccessTTL,
		RefreshTTL:   cfg.JWT.RefreshTTL,
		ResetBaseURL: getEnv("RESET_PASSWORD_URL", "http://localhost:5173/reset-password"),
	})

	// Admin password startup hook
	seedAdminPassword(ctx, userAuthRepo, cfg.AdminInitialPassword)

	// Public auth routes (before JWT middleware)
	authHandler := auth.NewHandler(authService, rateLimiter)
	publicAPI := router.Group("/api/v1")
	authHandler.RegisterPublicRoutes(publicAPI)

	// JWT auth middleware for all protected routes
	router.Use(middleware.JWTAuth([]byte(jwtSecret)))

	// Protected auth routes
	protectedAPI := router.Group("/api/v1")
	authHandler.RegisterProtectedRoutes(protectedAPI)

	// Admin routes
	adminGroup := router.Group("/api/v1/admin")
	metadataHandler.RegisterRoutes(adminGroup)
	secHandler := handler.NewSecurityHandler(roleService, psService, profileService, userService, permissionService, groupService, sharingRuleService, authService)
	secHandler.RegisterRoutes(adminGroup)

	// App templates
	templateRegistry := templates.BuildRegistry()
	templateApplier := templates.NewApplier(objectService, fieldService, objectRepo, permissionService, metadataCache)
	templateHandler := handler.NewTemplateHandler(templateRegistry, templateApplier)
	templateHandler.RegisterRoutes(adminGroup)

	// Territory management (enterprise only, no-op in community build)
	registerTerritoryRoutes(pool, adminGroup)

	// --- Security enforcers ---
	effectivePermRepo := effectiveRepo
	rlsCacheRepo := security.NewPgRLSEffectiveCacheRepository(pool)
	olsEnforcer := ols.NewEnforcer(effectivePermRepo)
	flsEnforcer := fls.NewEnforcer(effectivePermRepo)
	rlsMetadataAdapter := security.NewPgMetadataFieldLister(pool)
	rlsEnforcer := rls.NewEnforcer(rlsCacheRepo, rlsMetadataAdapter)

	// --- SOQL engine ---
	soqlMetadataAdapter := soql.NewMetadataAdapter(metadataCache)
	soqlAccessAdapter := soql.NewAccessControllerAdapter(metadataCache, olsEnforcer, flsEnforcer)
	soqlEngine := soqlengine.NewEngine(
		soqlengine.WithMetadata(soqlMetadataAdapter),
		soqlengine.WithAccessController(soqlAccessAdapter),
	)
	soqlExecutor := soql.NewExecutor(pool, metadataCache, rlsEnforcer)
	soqlService := soql.NewQueryService(soqlEngine, soqlExecutor)

	// --- DML engine ---
	dmlMetadataAdapter := dml.NewMetadataAdapter(metadataCache)
	dmlAccessAdapter := dml.NewWriteAccessControllerAdapter(metadataCache, olsEnforcer, flsEnforcer)
	dmlEngine := dmlengine.NewEngine(
		dmlengine.WithMetadata(dmlMetadataAdapter),
		dmlengine.WithWriteAccessController(dmlAccessAdapter),
	)
	dmlExecutor := dml.NewRLSExecutor(pool, metadataCache, rlsEnforcer)
	dmlService := dml.NewDMLService(dmlEngine, dmlExecutor)

	// --- Query/Data API ---
	queryHandler := handler.NewQueryHandler(soqlService, dmlService)
	apiGroup := router.Group("/api/v1")
	queryHandler.RegisterRoutes(apiGroup)

	return router
}

func startOutboxWorker(ctx context.Context, pool *pgxpool.Pool, dsn string, logger *slog.Logger) {
	connConfig, err := security.ParseConnConfig(dsn)
	if err != nil {
		slog.Error("failed to parse conn config for outbox worker", "error", err)
		return
	}

	outboxRepo := security.NewPgOutboxRepository(pool)
	userRepo := security.NewPgUserRepository(pool)
	profileRepo := security.NewPgProfileRepository(pool)
	psToUserRepo := security.NewPgPermissionSetToUserRepository(pool)
	psRepo := security.NewPgPermissionSetRepository(pool)
	objPermRepo := security.NewPgObjectPermissionRepository(pool)
	fieldPermRepo := security.NewPgFieldPermissionRepository(pool)
	effectiveRepo := security.NewPgEffectivePermissionRepository(pool)
	metadataLister := security.NewPgMetadataFieldLister(pool)

	computer := security.NewEffectiveComputer(
		pool, userRepo, profileRepo, psToUserRepo, psRepo,
		objPermRepo, fieldPermRepo, effectiveRepo, metadataLister,
	)

	// RLS effective computer
	userRoleRepo := security.NewPgUserRoleRepository(pool)
	groupRepo := security.NewPgGroupRepository(pool)
	memberRepo := security.NewPgGroupMemberRepository(pool)
	rlsCacheRepo := security.NewPgRLSEffectiveCacheRepository(pool)

	rlsComputer := security.NewRLSEffectiveComputer(
		pool, userRoleRepo, userRepo, groupRepo, memberRepo,
		rlsCacheRepo, metadataLister,
	)

	worker := security.NewOutboxWorker(*connConfig, outboxRepo, computer, rlsComputer, logger)
	go func() {
		if err := worker.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("outbox worker stopped with error", "error", err)
		}
	}()

	slog.Info("outbox worker started")
}

func seedAdminPassword(ctx context.Context, userAuthRepo auth.UserAuthRepository, password string) {
	if password == "" {
		return
	}

	user, err := userAuthRepo.GetByID(ctx, security.SystemAdminUserID)
	if err != nil {
		slog.Error("failed to load admin user for password seeding", "error", err)
		return
	}
	if user == nil {
		slog.Warn("admin user not found, skipping password seed")
		return
	}
	if user.PasswordHash != "" {
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("failed to hash admin password", "error", err)
		return
	}

	if err := userAuthRepo.SetPassword(ctx, security.SystemAdminUserID, hash); err != nil {
		slog.Error("failed to set admin password", "error", err)
		return
	}

	slog.Info("admin initial password set from ADMIN_INITIAL_PASSWORD")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
