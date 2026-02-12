// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

export type ModelStatus = 'planning' | 'active' | 'archived'

export interface TerritoryModel {
  id: string
  apiName: string
  label: string
  description: string
  status: ModelStatus
  activatedAt: string | null
  archivedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateModelRequest {
  apiName: string
  label: string
  description?: string
}

export interface UpdateModelRequest {
  label: string
  description?: string
}

export interface Territory {
  id: string
  modelId: string
  parentId: string | null
  apiName: string
  label: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface CreateTerritoryRequest {
  modelId: string
  parentId?: string | null
  apiName: string
  label: string
  description?: string
}

export interface UpdateTerritoryRequest {
  parentId?: string | null
  label: string
  description?: string
}

export interface TerritoryObjectDefault {
  id: string
  territoryId: string
  objectId: string
  accessLevel: string
  createdAt: string
  updatedAt: string
}

export interface SetObjectDefaultRequest {
  objectId: string
  accessLevel: string
}

export interface UserTerritoryAssignment {
  id: string
  userId: string
  territoryId: string
  createdAt: string
}

export interface AssignUserRequest {
  userId: string
}

export interface RecordTerritoryAssignment {
  id: string
  recordId: string
  objectId: string
  territoryId: string
  reason: string
  createdAt: string
}

export interface AssignRecordRequest {
  recordId: string
  objectId: string
  reason?: string
}

export interface AssignmentRule {
  id: string
  territoryId: string
  objectId: string
  isActive: boolean
  ruleOrder: number
  criteriaField: string
  criteriaOp: string
  criteriaValue: string
  createdAt: string
  updatedAt: string
}

export interface CreateAssignmentRuleRequest {
  territoryId: string
  objectId: string
  isActive: boolean
  ruleOrder: number
  criteriaField: string
  criteriaOp: string
  criteriaValue: string
}

export interface UpdateAssignmentRuleRequest {
  isActive: boolean
  ruleOrder: number
  criteriaField: string
  criteriaOp: string
  criteriaValue: string
}

export interface ModelFilter {
  page?: number
  perPage?: number
}

export interface TerritoryFilter {
  modelId: string
  page?: number
  perPage?: number
}

export interface AssignmentRuleFilter {
  territoryId?: string
  page?: number
  perPage?: number
}

export const MODEL_STATUS_LABELS: Record<ModelStatus, string> = {
  planning: 'Планирование',
  active: 'Активна',
  archived: 'Архив',
}
