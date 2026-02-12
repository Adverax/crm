// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

import { ref } from 'vue'
import { defineStore } from 'pinia'
import { territoryApi } from '@/api/territory'
import type {
  TerritoryModel,
  CreateModelRequest,
  UpdateModelRequest,
  ModelFilter,
  Territory,
  CreateTerritoryRequest,
  UpdateTerritoryRequest,
  TerritoryFilter,
  TerritoryObjectDefault,
  SetObjectDefaultRequest,
  UserTerritoryAssignment,
  AssignUserRequest,
  AssignmentRule,
  CreateAssignmentRuleRequest,
  UpdateAssignmentRuleRequest,
  AssignmentRuleFilter,
} from '@/types/territory'
import type { PaginationMeta } from '@/types/metadata'

export const useTerritoryAdminStore = defineStore('territoryAdmin', () => {
  // Models
  const models = ref<TerritoryModel[]>([])
  const currentModel = ref<TerritoryModel | null>(null)
  const modelsPagination = ref<PaginationMeta | null>(null)
  const modelsLoading = ref(false)
  const modelsError = ref<string | null>(null)

  // Territories
  const territories = ref<Territory[]>([])
  const currentTerritory = ref<Territory | null>(null)
  const territoriesPagination = ref<PaginationMeta | null>(null)
  const territoriesLoading = ref(false)
  const territoriesError = ref<string | null>(null)

  // Object Defaults
  const objectDefaults = ref<TerritoryObjectDefault[]>([])

  // User Assignments
  const userAssignments = ref<UserTerritoryAssignment[]>([])

  // Assignment Rules
  const assignmentRules = ref<AssignmentRule[]>([])
  const currentAssignmentRule = ref<AssignmentRule | null>(null)
  const assignmentRulesPagination = ref<PaginationMeta | null>(null)
  const assignmentRulesLoading = ref(false)
  const assignmentRulesError = ref<string | null>(null)

  // --- Models ---

  async function fetchModels(filter?: ModelFilter) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      const response = await territoryApi.listModels(filter)
      models.value = response.data ?? []
      modelsPagination.value = response.pagination
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка загрузки моделей'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function fetchModel(modelId: string) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      const response = await territoryApi.getModel(modelId)
      currentModel.value = response.data
      return response.data
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка загрузки модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function createModel(data: CreateModelRequest) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      const response = await territoryApi.createModel(data)
      return response.data
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка создания модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function updateModel(modelId: string, data: UpdateModelRequest) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      const response = await territoryApi.updateModel(modelId, data)
      currentModel.value = response.data
      return response.data
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка обновления модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function deleteModel(modelId: string) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      await territoryApi.deleteModel(modelId)
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка удаления модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function activateModel(modelId: string) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      await territoryApi.activateModel(modelId)
      if (currentModel.value?.id === modelId) {
        await fetchModel(modelId)
      }
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка активации модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  async function archiveModel(modelId: string) {
    modelsLoading.value = true
    modelsError.value = null
    try {
      await territoryApi.archiveModel(modelId)
      if (currentModel.value?.id === modelId) {
        await fetchModel(modelId)
      }
    } catch (err) {
      modelsError.value = err instanceof Error ? err.message : 'Ошибка архивации модели'
      throw err
    } finally {
      modelsLoading.value = false
    }
  }

  // --- Territories ---

  async function fetchTerritories(filter: TerritoryFilter) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.listTerritories(filter)
      territories.value = response.data ?? []
      territoriesPagination.value = response.pagination
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка загрузки территорий'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function fetchTerritory(territoryId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.getTerritory(territoryId)
      currentTerritory.value = response.data
      return response.data
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка загрузки территории'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function createTerritory(data: CreateTerritoryRequest) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.createTerritory(data)
      return response.data
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка создания территории'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function updateTerritory(territoryId: string, data: UpdateTerritoryRequest) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.updateTerritory(territoryId, data)
      currentTerritory.value = response.data
      return response.data
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка обновления территории'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function deleteTerritory(territoryId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      await territoryApi.deleteTerritory(territoryId)
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка удаления территории'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  // --- Object Defaults ---

  async function fetchObjectDefaults(territoryId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.listObjectDefaults(territoryId)
      objectDefaults.value = response.data ?? []
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка загрузки настроек объектов'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function setObjectDefault(territoryId: string, data: SetObjectDefaultRequest) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.setObjectDefault(territoryId, data)
      return response.data
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка настройки объекта'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function removeObjectDefault(territoryId: string, objectId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      await territoryApi.removeObjectDefault(territoryId, objectId)
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка удаления настройки объекта'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  // --- User Assignments ---

  async function fetchTerritoryUsers(territoryId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.listTerritoryUsers(territoryId)
      userAssignments.value = response.data ?? []
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка загрузки пользователей'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function assignUser(territoryId: string, data: AssignUserRequest) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      const response = await territoryApi.assignUser(territoryId, data)
      return response.data
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка назначения пользователя'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  async function unassignUser(territoryId: string, userId: string) {
    territoriesLoading.value = true
    territoriesError.value = null
    try {
      await territoryApi.unassignUser(territoryId, userId)
    } catch (err) {
      territoriesError.value = err instanceof Error ? err.message : 'Ошибка удаления пользователя'
      throw err
    } finally {
      territoriesLoading.value = false
    }
  }

  // --- Assignment Rules ---

  async function fetchAssignmentRules(filter?: AssignmentRuleFilter) {
    assignmentRulesLoading.value = true
    assignmentRulesError.value = null
    try {
      const response = await territoryApi.listAssignmentRules(filter)
      assignmentRules.value = response.data ?? []
      assignmentRulesPagination.value = response.pagination
    } catch (err) {
      assignmentRulesError.value = err instanceof Error ? err.message : 'Ошибка загрузки правил'
      throw err
    } finally {
      assignmentRulesLoading.value = false
    }
  }

  async function fetchAssignmentRule(ruleId: string) {
    assignmentRulesLoading.value = true
    assignmentRulesError.value = null
    try {
      const response = await territoryApi.getAssignmentRule(ruleId)
      currentAssignmentRule.value = response.data
      return response.data
    } catch (err) {
      assignmentRulesError.value = err instanceof Error ? err.message : 'Ошибка загрузки правила'
      throw err
    } finally {
      assignmentRulesLoading.value = false
    }
  }

  async function createAssignmentRule(data: CreateAssignmentRuleRequest) {
    assignmentRulesLoading.value = true
    assignmentRulesError.value = null
    try {
      const response = await territoryApi.createAssignmentRule(data)
      return response.data
    } catch (err) {
      assignmentRulesError.value = err instanceof Error ? err.message : 'Ошибка создания правила'
      throw err
    } finally {
      assignmentRulesLoading.value = false
    }
  }

  async function updateAssignmentRule(ruleId: string, data: UpdateAssignmentRuleRequest) {
    assignmentRulesLoading.value = true
    assignmentRulesError.value = null
    try {
      const response = await territoryApi.updateAssignmentRule(ruleId, data)
      currentAssignmentRule.value = response.data
      return response.data
    } catch (err) {
      assignmentRulesError.value = err instanceof Error ? err.message : 'Ошибка обновления правила'
      throw err
    } finally {
      assignmentRulesLoading.value = false
    }
  }

  async function deleteAssignmentRule(ruleId: string) {
    assignmentRulesLoading.value = true
    assignmentRulesError.value = null
    try {
      await territoryApi.deleteAssignmentRule(ruleId)
    } catch (err) {
      assignmentRulesError.value = err instanceof Error ? err.message : 'Ошибка удаления правила'
      throw err
    } finally {
      assignmentRulesLoading.value = false
    }
  }

  return {
    // Models
    models,
    currentModel,
    modelsPagination,
    modelsLoading,
    modelsError,
    fetchModels,
    fetchModel,
    createModel,
    updateModel,
    deleteModel,
    activateModel,
    archiveModel,

    // Territories
    territories,
    currentTerritory,
    territoriesPagination,
    territoriesLoading,
    territoriesError,
    fetchTerritories,
    fetchTerritory,
    createTerritory,
    updateTerritory,
    deleteTerritory,

    // Object Defaults
    objectDefaults,
    fetchObjectDefaults,
    setObjectDefault,
    removeObjectDefault,

    // User Assignments
    userAssignments,
    fetchTerritoryUsers,
    assignUser,
    unassignUser,

    // Assignment Rules
    assignmentRules,
    currentAssignmentRule,
    assignmentRulesPagination,
    assignmentRulesLoading,
    assignmentRulesError,
    fetchAssignmentRules,
    fetchAssignmentRule,
    createAssignmentRule,
    updateAssignmentRule,
    deleteAssignmentRule,
  }
})
