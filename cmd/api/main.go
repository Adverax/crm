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

	"github.com/adverax/crm/internal/api"
	"github.com/adverax/crm/internal/handler"
	"github.com/adverax/crm/internal/middleware"
	"github.com/adverax/crm/internal/pkg/config"
	"github.com/adverax/crm/internal/pkg/database"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/metadata/ddl"
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

	// Repository adapters (will be replaced with real sqlc adapters)
	objectRepo := metadata.NewPgObjectRepository(pool)
	fieldRepo := metadata.NewPgFieldRepository(pool)
	polyRepo := metadata.NewPgPolymorphicTargetRepository(pool)
	ddlExec := ddl.NewExecutor()

	// Cache
	cacheLoader := metadata.NewPgCacheLoader(pool)
	metadataCache := metadata.NewMetadataCache(cacheLoader)

	// Load cache on startup
	ctx := context.Background()
	if err := metadataCache.Load(ctx); err != nil {
		slog.Warn("metadata cache initial load failed (empty database?)", "error", err)
	}

	// Services
	objectService := metadata.NewObjectService(pool, objectRepo, fieldRepo, ddlExec, metadataCache)
	fieldService := metadata.NewFieldService(pool, objectRepo, fieldRepo, polyRepo, ddlExec, metadataCache)

	// Handler
	metadataHandler := handler.NewMetadataHandler(objectService, fieldService)

	// Register generated routes
	api.RegisterHandlers(router, metadataHandler)

	return router
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
