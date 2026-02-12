//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package setup

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/ee/internal/handler"
	"github.com/adverax/crm/ee/internal/platform/territory"
)

// RegisterTerritoryRoutes instantiates all territory services and registers routes.
func RegisterTerritoryRoutes(pool *pgxpool.Pool, adminGroup *gin.RouterGroup) {
	modelRepo := territory.NewPgModelRepository(pool)
	territoryRepo := territory.NewPgTerritoryRepository(pool)
	objDefaultRepo := territory.NewPgObjectDefaultRepository(pool)
	userAssignmentRepo := territory.NewPgUserAssignmentRepository(pool)
	recordAssignmentRepo := territory.NewPgRecordAssignmentRepository(pool)
	assignmentRuleRepo := territory.NewPgAssignmentRuleRepository(pool)
	effectiveRepo := territory.NewPgEffectiveRepository(pool)
	objLookup := territory.NewPgObjectDefinitionLookup(pool)

	modelService := territory.NewModelService(pool, modelRepo, effectiveRepo)
	territoryService := territory.NewTerritoryService(pool, territoryRepo, modelRepo)
	objDefaultService := territory.NewObjectDefaultService(pool, objDefaultRepo, territoryRepo)
	userAssignService := territory.NewUserAssignmentService(pool, userAssignmentRepo, territoryRepo)
	recAssignService := territory.NewRecordAssignmentService(pool, recordAssignmentRepo, territoryRepo, effectiveRepo, objLookup)
	ruleService := territory.NewAssignmentRuleService(pool, assignmentRuleRepo, territoryRepo)

	terrHandler := handler.NewTerritoryHandler(
		modelService,
		territoryService,
		objDefaultService,
		userAssignService,
		recAssignService,
		ruleService,
	)
	terrHandler.RegisterRoutes(adminGroup)
}
