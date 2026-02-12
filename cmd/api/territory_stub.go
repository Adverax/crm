//go:build !enterprise

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerTerritoryRoutes(_ *pgxpool.Pool, _ *gin.RouterGroup) {
	// Territory management is an enterprise feature.
	// Build with -tags enterprise to enable.
}
