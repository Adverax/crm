<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus, GripVertical } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { VueDraggable } from 'vue-draggable-plus'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import GateBodyStepInstructionEditor from './GateBodyStepInstructionEditor.vue'
import { ref, computed } from 'vue'
import type { GateBodyStep } from '@/types/portals'

const props = defineProps<{
  steps: GateBodyStep[]
}>()

const emit = defineEmits<{
  'update:steps': [value: GateBodyStep[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedStep = computed(() =>
  selectedIndex.value !== null ? props.steps[selectedIndex.value] ?? null : null,
)

function selectStep(index: number) {
  selectedIndex.value = index
}

function addStep() {
  const updated: GateBodyStep[] = [...props.steps, { name: '', type: 'soql', soql: '' }]
  emit('update:steps', updated)
  selectedIndex.value = updated.length - 1
}

function removeStep(index: number) {
  const updated = [...props.steps]
  updated.splice(index, 1)
  emit('update:steps', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

function onDragEnd() {
  emit('update:steps', [...localSteps.value])
}

const localSteps = computed({
  get: () => props.steps,
  set: (val) => emit('update:steps', [...val]),
})

function detectStepType(text: string): 'soql' | 'dml' | null {
  const stripped = text
    .split('\n')
    .filter((line) => !line.trimStart().startsWith('--'))
    .join('\n')
    .trim()
  if (!stripped) return null
  const firstWord = (stripped.split(/\s/)[0] ?? '').toUpperCase()
  if (firstWord === 'SELECT') return 'soql'
  if (['INSERT', 'UPDATE', 'DELETE', 'UPSERT'].includes(firstWord)) return 'dml'
  return null
}

const stepText = computed(() => {
  if (!selectedStep.value) return ''
  return selectedStep.value.type === 'dml'
    ? selectedStep.value.dml ?? ''
    : selectedStep.value.soql ?? ''
})

const detectedType = computed(() => detectStepType(stepText.value))

function onInstructionChange(text: string) {
  if (!selectedStep.value) return
  const detected = detectStepType(text)
  const resolvedType = detected ?? 'soql'
  selectedStep.value.type = resolvedType
  if (resolvedType === 'dml') {
    selectedStep.value.dml = text
    selectedStep.value.soql = undefined
    selectedStep.value.pageSize = undefined
    selectedStep.value.restrictFilters = undefined
  } else {
    selectedStep.value.soql = text
    selectedStep.value.dml = undefined
  }
}
</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="body-steps-editor">
    <!-- Left panel: step list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Instructions</span>
        <IconButton
          :icon="Plus"
          tooltip="Add step"
          variant="outline"
          size="sm"
          data-testid="add-step-btn"
          @click="addStep"
        />
      </div>
      <div v-if="steps.length === 0" class="p-3 text-sm text-muted-foreground">
        No instructions configured.
      </div>
      <VueDraggable
        v-else
        v-model="localSteps"
        :animation="150"
        handle=".drag-handle"
        class="divide-y"
        @end="onDragEnd"
      >
        <button
          v-for="(step, sIdx) in steps"
          :key="sIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedIndex === sIdx }"
          data-testid="step-card"
          @click="selectStep(sIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <GripVertical class="h-4 w-4 text-muted-foreground drag-handle cursor-grab shrink-0" />
            <div class="min-w-0 flex-1">
              <div class="text-sm font-medium font-mono truncate">{{ step.name || 'unnamed' }}</div>
            </div>
            <Badge variant="secondary" class="shrink-0 text-[10px]">
              {{ step.type }}
            </Badge>
          </div>
        </button>
      </VueDraggable>
    </div>

    <!-- Right panel: step detail -->
    <div class="flex-1 min-w-0">
      <div
        v-if="!selectedStep"
        class="flex items-center justify-center h-full text-sm text-muted-foreground"
      >
        Select a step to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Instruction</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete step"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-step-btn"
            @click="removeStep(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input
            v-model="selectedStep.name"
            placeholder="step_name"
            class="font-mono"
            data-testid="step-name-input"
          />
        </div>

        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <Label class="text-xs">Instruction</Label>
            <Badge variant="secondary" class="text-[10px]" data-testid="step-type-badge">
              {{ detectedType ?? 'auto' }}
            </Badge>
          </div>
          <GateBodyStepInstructionEditor
            :model-value="stepText"
            :detected-type="detectedType"
            height="120px"
            @update:model-value="onInstructionChange"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">When (CEL)</Label>
          <p class="text-[11px] text-muted-foreground">
            Optional condition. Step is skipped when false.
          </p>
          <ExpressionBuilder
            :model-value="selectedStep.when ?? ''"
            context="gate_when"
            height="80px"
            placeholder="size(args.ids) > 0"
            :show-field-picker="false"
            data-testid="step-when-input"
            @update:model-value="selectedStep.when = $event || undefined"
          />
        </div>

        <div v-if="detectedType === 'soql'" class="space-y-1">
          <Label class="text-xs">Page Size</Label>
          <p class="text-[11px] text-muted-foreground">
            Items per page for paginated queries. Leave empty for default (200).
          </p>
          <Input
            :model-value="selectedStep.pageSize ?? ''"
            type="number"
            min="1"
            max="200"
            placeholder="200"
            data-testid="step-page-size-input"
            @update:model-value="selectedStep.pageSize = $event ? Number($event) : undefined"
          />
        </div>

        <div v-if="detectedType === 'soql'" class="space-y-1">
          <Label class="text-xs">Restrict Filters</Label>
          <p class="text-[11px] text-muted-foreground">
            Fields to exclude from dynamic filtering. Comma-separated.
          </p>
          <Input
            :model-value="(selectedStep.restrictFilters ?? []).join(', ')"
            placeholder="e.g. OwnerId, Status"
            data-testid="step-restrict-filters-input"
            @update:model-value="selectedStep.restrictFilters = ($event as string).split(',').map((s: string) => s.trim()).filter(Boolean)"
          />
        </div>
      </div>
    </div>
  </div>
</template>
