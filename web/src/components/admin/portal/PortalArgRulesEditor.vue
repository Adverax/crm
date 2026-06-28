<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import { ref, computed } from 'vue'
import type { PortalArgRule } from '@/types/portals'

const props = defineProps<{
  rules: PortalArgRule[]
}>()

const emit = defineEmits<{
  'update:rules': [value: PortalArgRule[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedRule = computed(() =>
  selectedIndex.value !== null ? props.rules[selectedIndex.value] ?? null : null,
)

const nameRegex = /^[a-z][a-z0-9_]*$/

// --- Validation ---

const nameError = computed<string | null>(() => {
  if (selectedIndex.value === null || !selectedRule.value) return null
  const name = selectedRule.value.name
  if (!name) return 'Name is required'
  if (!nameRegex.test(name)) return 'Must match ^[a-z][a-z0-9_]*$'
  const isDuplicate = props.rules.some(
    (r, i) => i !== selectedIndex.value && r.name === name,
  )
  if (isDuplicate) return `Duplicate name "${name}"`
  return null
})

const conditionError = computed<string | null>(() => {
  if (!selectedRule.value) return null
  if (!selectedRule.value.condition) return 'Condition is required'
  return null
})

const errorMessageError = computed<string | null>(() => {
  if (!selectedRule.value) return null
  if (!selectedRule.value.errorMessage) return 'Error message is required'
  return null
})

const hasErrors = computed(() => {
  for (let i = 0; i < props.rules.length; i++) {
    if (hasRuleError(i)) return true
  }
  return false
})

function hasRuleError(index: number): boolean {
  const r = props.rules[index]
  if (!r) return false
  if (!r.name || !nameRegex.test(r.name)) return true
  if (props.rules.some((b, j) => j !== index && b.name === r.name)) return true
  if (!r.condition) return true
  if (!r.errorMessage) return true
  return false
}

defineExpose({ hasErrors })

// --- Actions ---

function selectRule(index: number) {
  selectedIndex.value = index
}

function addRule() {
  const updated: PortalArgRule[] = [...props.rules, { name: '', condition: '', errorMessage: '' }]
  emit('update:rules', updated)
  selectedIndex.value = updated.length - 1
}

function removeRule(index: number) {
  const updated = [...props.rules]
  updated.splice(index, 1)
  emit('update:rules', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

function updateCondition(value: string) {
  if (!selectedRule.value) return
  selectedRule.value.condition = value
}

function updateErrorMessage(value: string) {
  if (!selectedRule.value) return
  selectedRule.value.errorMessage = value
}
</script>

<template>
  <div class="flex gap-4 min-h-[300px]" data-testid="arg-rules-editor">
    <!-- Left panel: rule list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Rules</span>
        <IconButton
          :icon="Plus"
          tooltip="Add rule"
          variant="outline"
          size="sm"
          data-testid="add-arg-rule-btn"
          @click="addRule"
        />
      </div>
      <div v-if="rules.length === 0" class="p-3 text-sm text-muted-foreground">
        No arg rules configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(rule, rIdx) in rules"
          :key="rIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{
            'bg-muted': selectedIndex === rIdx,
            'border-l-2 border-l-destructive': hasRuleError(rIdx),
          }"
          data-testid="arg-rule-card"
          @click="selectRule(rIdx)"
        >
          <div class="text-sm font-medium font-mono truncate">{{ rule.name || 'unnamed' }}</div>
        </button>
      </div>
    </div>

    <!-- Right panel: rule detail -->
    <div class="flex-1 min-w-0">
      <div
        v-if="!selectedRule"
        class="flex items-center justify-center h-full text-sm text-muted-foreground"
      >
        Select a rule to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Rule Details</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete rule"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-arg-rule-btn"
            @click="removeRule(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input
            v-model="selectedRule.name"
            placeholder="rule_name"
            class="font-mono"
            :class="{ 'border-destructive': nameError }"
            data-testid="arg-rule-name-input"
          />
          <p v-if="nameError" class="text-[11px] text-destructive">
            {{ nameError }}
          </p>
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Condition (CEL)</Label>
          <p class="text-[11px] text-muted-foreground">
            CEL expression evaluated at runtime. Must return true for the request to proceed.
            Use <span class="font-mono">args.*</span> to reference arg values.
          </p>
          <ExpressionBuilder
            :model-value="selectedRule.condition"
            context="portal_when"
            height="80px"
            placeholder="args.limit > 0 && args.limit <= 100"
            :show-field-picker="false"
            :class="{ 'border-destructive': conditionError }"
            data-testid="arg-rule-condition-input"
            @update:model-value="updateCondition($event)"
          />
          <p v-if="conditionError" class="text-[11px] text-destructive">
            {{ conditionError }}
          </p>
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Error Message</Label>
          <p class="text-[11px] text-muted-foreground">
            Message returned to the client when condition evaluates to false.
          </p>
          <Input
            :model-value="selectedRule.errorMessage"
            placeholder="Limit must be between 1 and 100"
            :class="{ 'border-destructive': errorMessageError }"
            data-testid="arg-rule-error-message-input"
            @update:model-value="updateErrorMessage(String($event))"
          />
          <p v-if="errorMessageError" class="text-[11px] text-destructive">
            {{ errorMessageError }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
