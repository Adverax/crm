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
	"github.com/adverax/crm/internal/pkg/config"
	"github.com/adverax/crm/internal/pkg/database"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/metadata/ddl"
	"github.com/adverax/crm/internal/platform/security"
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

	router := setupRouter(pool)

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

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
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

	// Dev auth middleware (replaces JWT in Phase 5)
	router.Use(middleware.DevAuth(userRepo, security.SystemAdminUserID))

	// Admin routes
	adminGroup := router.Group("/api/v1/admin")
	metadataHandler.RegisterRoutes(adminGroup)
	secHandler := handler.NewSecurityHandler(roleService, psService, profileService, userService, permissionService, groupService, sharingRuleService)
	secHandler.RegisterRoutes(adminGroup)

	// Effective permission computer (used by outbox worker)
	_ = effectiveRepo

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
