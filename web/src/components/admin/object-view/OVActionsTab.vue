<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ref, computed } from 'vue'
import type { OVAction, OVActionApply } from '@/types/object-views'

const props = defineProps<{
  actions: OVAction[]
}>()

const emit = defineEmits<{
  'update:actions': [value: OVAction[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedAction = computed(() =>
  selectedIndex.value !== null ? props.actions[selectedIndex.value] ?? null : null,
)

function selectAction(index: number) {
  selectedIndex.value = index
}

function addAction() {
  const updated = [...props.actions, {
    key: `action_${Date.now()}`,
    label: 'New Action',
    type: 'secondary',
    icon: '',
    visibilityExpr: '',
  }]
  emit('update:actions', updated)
  selectedIndex.value = updated.length - 1
}

function removeAction(index: number) {
  const updated = [...props.actions]
  updated.splice(index, 1)
  emit('update:actions', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onActionTypeChange(value: any) {
  if (selectedAction.value) selectedAction.value.type = String(value)
}

// --- Form Fields ---
function addFormField(action: OVAction) {
  if (!action.form) action.form = []
  action.form.push({ name: '', type: 'string', label: '', required: false, default: '' })
}

function removeFormField(action: OVAction, idx: number) {
  action.form?.splice(idx, 1)
}

// --- Validation ---
function addValidation(action: OVAction) {
  if (!action.validation) action.validation = []
  action.validation.push({ expr: '', message: '', code: '' })
}

function removeValidation(action: OVAction, idx: number) {
  action.validation?.splice(idx, 1)
}

// --- Apply ---
function ensureApply(action: OVAction): OVActionApply {
  if (!action.apply) {
    action.apply = { type: 'dml', dml: [] }
  }
  return action.apply
}

function removeApply(action: OVAction) {
  action.apply = undefined
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onApplyTypeChange(action: OVAction, value: any) {
  const apply = ensureApply(action)
  apply.type = String(value) as 'dml' | 'scenario'
}

function addDmlStatement(apply: OVActionApply) {
  if (!apply.dml) apply.dml = []
  apply.dml.push('')
}

function removeDmlStatement(apply: OVActionApply, idx: number) {
  apply.dml?.splice(idx, 1)
}

function typeBadgeVariant(type: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (type === 'primary') return 'default'
  if (type === 'danger') return 'destructive'
  return 'secondary'
}
</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="actions-master-detail">
    <!-- Left panel: action list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Actions</span>
        <IconButton
          :icon="Plus"
          tooltip="Add action"
          variant="outline"
          size="sm"
          data-testid="add-action-btn"
          @click="addAction"
        />
      </div>
      <div v-if="actions.length === 0" class="p-3 text-sm text-muted-foreground">
        No actions configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(action, aIdx) in actions"
          :key="aIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedIndex === aIdx }"
          data-testid="action-card"
          @click="selectAction(aIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium truncate">{{ action.label || action.key }}</div>
              <div class="text-xs text-muted-foreground font-mono truncate">{{ action.key }}</div>
            </div>
            <Badge :variant="typeBadgeVariant(action.type)" class="shrink-0 text-[10px]">
              {{ action.type }}
            </Badge>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: action detail with tabs -->
    <div class="flex-1 min-w-0">
      <div v-if="!selectedAction" class="flex items-center justify-center h-full text-sm text-muted-foreground">
        Select an action to edit
      </div>

      <Tabs v-else default-value="identity" class="h-full">
        <div class="flex items-center justify-between mb-3">
          <TabsList>
            <TabsTrigger value="identity" data-testid="tab-action-identity">Identity</TabsTrigger>
            <TabsTrigger value="form" data-testid="tab-action-form">
              Form ({{ selectedAction.form?.length ?? 0 }})
            </TabsTrigger>
            <TabsTrigger value="validation" data-testid="tab-action-validation">
              Validation ({{ selectedAction.validation?.length ?? 0 }})
            </TabsTrigger>
            <TabsTrigger value="apply" data-testid="tab-action-apply">
              Apply {{ selectedAction.apply ? `(${selectedAction.apply.type})` : '' }}
            </TabsTrigger>
          </TabsList>
          <IconButton
            :icon="Trash2"
            tooltip="Delete action"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-action-btn"
            @click="removeAction(selectedIndex!)"
          />
        </div>

        <!-- Identity tab -->
        <TabsContent value="identity" class="space-y-4" data-testid="action-identity-tab">
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <Label class="text-xs">Key</Label>
              <Input v-model="selectedAction.key" placeholder="action_key" class="font-mono" />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">Label</Label>
              <Input v-model="selectedAction.label" placeholder="Action Label" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <Label class="text-xs">Type</Label>
              <Select
                :model-value="selectedAction.type"
                @update:model-value="onActionTypeChange"
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="primary">Primary</SelectItem>
                  <SelectItem value="secondary">Secondary</SelectItem>
                  <SelectItem value="danger">Danger</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-1">
              <Label class="text-xs">Icon (lucide name)</Label>
              <Input v-model="selectedAction.icon" placeholder="mail" class="font-mono" />
            </div>
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Visibility Expression (CEL)</Label>
            <ExpressionBuilder
              v-model="selectedAction.visibilityExpr"
              context="when_expression"
              height="80px"
              placeholder="record.status == 'draft'"
            />
          </div>
        </TabsContent>

        <!-- Form tab -->
        <TabsContent value="form" class="space-y-3" data-testid="action-form-tab">
          <div
            v-for="(field, fIdx) in (selectedAction.form ?? [])"
            :key="fIdx"
            class="grid grid-cols-5 gap-2 items-end"
            data-testid="action-form-field"
          >
            <div class="space-y-1">
              <Label class="text-xs">Name</Label>
              <Input v-model="field.name" placeholder="field_name" class="font-mono text-xs" />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">Type</Label>
              <Input v-model="field.type" placeholder="string" class="font-mono text-xs" />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">Label</Label>
              <Input v-model="field.label" placeholder="Field Label" class="text-xs" />
            </div>
            <div class="flex items-center gap-2">
              <Checkbox :checked="field.required" @update:checked="(v: boolean) => field.required = v" />
              <Label class="text-xs">Required</Label>
            </div>
            <div class="flex items-end gap-1">
              <div class="flex-1 space-y-1">
                <Label class="text-xs">Default</Label>
                <Input v-model="field.default" placeholder="'value'" class="font-mono text-xs" />
              </div>
              <IconButton
                :icon="Trash2"
                tooltip="Remove field"
                variant="ghost"
                size="sm"
                class="text-destructive hover:text-destructive"
                @click="removeFormField(selectedAction, fIdx)"
              />
            </div>
          </div>
          <div v-if="(selectedAction.form?.length ?? 0) === 0" class="text-sm text-muted-foreground">
            No form fields. This action won't show a dialog.
          </div>
          <IconButton
            :icon="Plus"
            tooltip="Add form field"
            variant="outline"
            size="sm"
            data-testid="add-form-field-btn"
            @click="addFormField(selectedAction)"
          />
        </TabsContent>

        <!-- Validation tab -->
        <TabsContent value="validation" class="space-y-3" data-testid="action-validation-tab">
          <div
            v-for="(rule, vIdx) in (selectedAction.validation ?? [])"
            :key="vIdx"
            class="space-y-2 border rounded-md p-3"
            data-testid="action-validation-rule"
          >
            <div class="flex items-center justify-between">
              <Label class="text-xs">Expression (CEL)</Label>
              <IconButton
                :icon="Trash2"
                tooltip="Remove rule"
                variant="ghost"
                size="sm"
                class="text-destructive hover:text-destructive"
                @click="removeValidation(selectedAction, vIdx)"
              />
            </div>
            <ExpressionBuilder
              v-model="rule.expr"
              context="validation_rule"
              height="80px"
              placeholder="size(data.name) > 0"
            />
            <div class="grid grid-cols-2 gap-2">
              <div class="space-y-1">
                <Label class="text-xs">Message</Label>
                <Input v-model="rule.message" placeholder="Name is required" class="text-xs" />
              </div>
              <div class="space-y-1">
                <Label class="text-xs">Code</Label>
                <Input v-model="rule.code" placeholder="name_required" class="font-mono text-xs" />
              </div>
            </div>
          </div>
          <div v-if="(selectedAction.validation?.length ?? 0) === 0" class="text-sm text-muted-foreground">
            No validation rules.
          </div>
          <IconButton
            :icon="Plus"
            tooltip="Add validation rule"
            variant="outline"
            size="sm"
            data-testid="add-validation-rule-btn"
            @click="addValidation(selectedAction)"
          />
        </TabsContent>

        <!-- Apply tab -->
        <TabsContent value="apply" class="space-y-3" data-testid="action-apply-tab">
          <div v-if="!selectedAction.apply" class="space-y-3">
            <div class="text-sm text-muted-foreground">
              No execution configured (UI-only action).
            </div>
            <IconButton
              :icon="Plus"
              tooltip="Add apply config"
              variant="outline"
              size="sm"
              data-testid="add-apply-btn"
              @click="ensureApply(selectedAction)"
            />
          </div>
          <template v-else>
            <div class="flex items-center gap-3">
              <div class="space-y-1 w-40">
                <Label class="text-xs">Type</Label>
                <Select
                  :model-value="selectedAction.apply.type"
                  @update:model-value="(v) => onApplyTypeChange(selectedAction!, v)"
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="dml">DML</SelectItem>
                    <SelectItem value="scenario">Scenario</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <IconButton
                :icon="Trash2"
                tooltip="Remove apply config"
                variant="ghost"
                size="sm"
                class="text-destructive hover:text-destructive mt-4"
                @click="removeApply(selectedAction)"
              />
            </div>

            <!-- DML statements -->
            <div v-if="selectedAction.apply.type === 'dml'" class="space-y-2">
              <div
                v-for="(stmt, dIdx) in (selectedAction.apply.dml ?? [])"
                :key="dIdx"
                class="flex gap-2 items-start"
                data-testid="dml-statement"
              >
                <Textarea
                  :model-value="stmt"
                  @update:model-value="(v) => { if (selectedAction!.apply?.dml) selectedAction!.apply.dml[dIdx] = String(v) }"
                  placeholder="INSERT INTO Account (Name) VALUES (data.Name)"
                  class="font-mono text-xs flex-1"
                  rows="2"
                />
                <IconButton
                  :icon="Trash2"
                  tooltip="Remove DML statement"
                  variant="ghost"
                  size="sm"
                  class="text-destructive hover:text-destructive mt-1"
                  @click="removeDmlStatement(selectedAction.apply!, dIdx)"
                />
              </div>
              <IconButton
                :icon="Plus"
                tooltip="Add DML statement"
                variant="outline"
                size="sm"
                data-testid="add-dml-btn"
                @click="addDmlStatement(selectedAction.apply!)"
              />
            </div>

            <!-- Scenario ref -->
            <div v-if="selectedAction.apply.type === 'scenario'" class="space-y-2">
              <div class="space-y-1">
                <Label class="text-xs">Scenario API Name</Label>
                <Input
                  :model-value="selectedAction.apply.scenario?.apiName ?? ''"
                  @update:model-value="(v) => { if (!selectedAction!.apply!.scenario) selectedAction!.apply!.scenario = { apiName: '' }; selectedAction!.apply!.scenario.apiName = String(v) }"
                  placeholder="scenario_api_name"
                  class="font-mono text-xs"
                />
              </div>
            </div>
          </template>
        </TabsContent>
      </Tabs>
    </div>
  </div>
</template>
