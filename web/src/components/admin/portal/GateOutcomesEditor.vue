<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ref, computed } from 'vue'
import type { GateOutcome } from '@/types/portals'

const props = defineProps<{
  outcomes: GateOutcome[]
  gateNames: string[]
}>()

const emit = defineEmits<{
  'update:outcomes': [value: GateOutcome[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedOutcome = computed(() =>
  selectedIndex.value !== null ? props.outcomes[selectedIndex.value] ?? null : null,
)

function selectOutcome(index: number) {
  selectedIndex.value = index
}

function addOutcome() {
  const updated: GateOutcome[] = [...props.outcomes, { name: '', gate: '' }]
  emit('update:outcomes', updated)
  selectedIndex.value = updated.length - 1
}

function removeOutcome(index: number) {
  const updated = [...props.outcomes]
  updated.splice(index, 1)
  emit('update:outcomes', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

const outcomeTypes = ['primary', 'secondary', 'danger', 'navigate'] as const
</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="outcomes-editor">
    <!-- Left panel: outcome list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Outcomes</span>
        <IconButton
          :icon="Plus"
          tooltip="Add outcome"
          variant="outline"
          size="sm"
          data-testid="add-outcome-btn"
          @click="addOutcome"
        />
      </div>
      <div v-if="outcomes.length === 0" class="p-3 text-sm text-muted-foreground">
        No outcomes configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(outcome, oIdx) in outcomes"
          :key="oIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedIndex === oIdx }"
          data-testid="outcome-card"
          @click="selectOutcome(oIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium font-mono truncate">{{ outcome.name || 'unnamed' }}</div>
              <div v-if="outcome.gate" class="text-[11px] text-muted-foreground truncate">
                &rarr; {{ outcome.gate }}
              </div>
            </div>
            <Badge v-if="outcome.type" variant="secondary" class="shrink-0 text-[10px]">
              {{ outcome.type }}
            </Badge>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: outcome detail -->
    <div class="flex-1 min-w-0">
      <div
        v-if="!selectedOutcome"
        class="flex items-center justify-center h-full text-sm text-muted-foreground"
      >
        Select an outcome to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Outcome Details</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete outcome"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-outcome-btn"
            @click="removeOutcome(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input
            v-model="selectedOutcome.name"
            placeholder="outcome_name"
            class="font-mono"
            data-testid="outcome-name-input"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Target Gate</Label>
          <Select v-model="selectedOutcome.gate" data-testid="outcome-gate-select">
            <SelectTrigger>
              <SelectValue placeholder="Select target gate" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="gn in gateNames" :key="gn" :value="gn">
                {{ gn }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Label</Label>
          <Input
            :model-value="selectedOutcome.label ?? ''"
            placeholder="View Details"
            data-testid="outcome-label-input"
            @update:model-value="selectedOutcome.label = String($event) || undefined"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Icon</Label>
          <Input
            :model-value="selectedOutcome.icon ?? ''"
            placeholder="eye"
            class="font-mono"
            data-testid="outcome-icon-input"
            @update:model-value="selectedOutcome.icon = String($event) || undefined"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Type</Label>
          <Select
            :model-value="selectedOutcome.type ?? ''"
            data-testid="outcome-type-select"
            @update:model-value="selectedOutcome.type = String($event) || undefined"
          >
            <SelectTrigger>
              <SelectValue placeholder="(none)" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">None</SelectItem>
              <SelectItem v-for="ot in outcomeTypes" :key="ot" :value="ot">
                {{ ot }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  </div>
</template>
