//go:build enterprise

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/ee/setup"
)

func registerTerritoryRoutes(pool *pgxpool.Pool, adminGroup *gin.RouterGroup) {
	setup.RegisterTerritoryRoutes(pool, adminGroup)
}
