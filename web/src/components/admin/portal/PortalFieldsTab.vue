<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import { ref, computed } from 'vue'
import type { PortalViewField, PortalQuery } from '@/types/portals'

const props = defineProps<{
  fields: PortalViewField[]
  queries: PortalQuery[]
}>()

const emit = defineEmits<{
  'update:fields': [value: PortalViewField[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedField = computed(() =>
  selectedIndex.value !== null ? props.fields[selectedIndex.value] ?? null : null,
)

function selectField(index: number) {
  selectedIndex.value = index
}

function addField() {
  const updated: PortalViewField[] = [...props.fields, { name: '' }]
  emit('update:fields', updated)
  selectedIndex.value = updated.length - 1
}

function removeField(index: number) {
  const updated = [...props.fields]
  updated.splice(index, 1)
  emit('update:fields', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

function fieldLabel(field: PortalViewField): string {
  return field.name || 'unnamed'
}

function isComputed(field: PortalViewField): boolean {
  return !!field.expr
}

function queryNames(): string[] {
  return props.queries.map((q) => q.name).filter(Boolean)
}
</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="fields-master-detail">
    <!-- Left panel: field list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Fields</span>
        <IconButton
          :icon="Plus"
          tooltip="Add field"
          variant="outline"
          size="sm"
          data-testid="add-field-btn"
          @click="addField"
        />
      </div>
      <p class="px-3 py-1.5 text-[11px] text-muted-foreground border-b">
        First 3 fields are used as highlights
      </p>
      <div v-if="fields.length === 0" class="p-3 text-sm text-muted-foreground">
        No fields configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(field, fIdx) in fields"
          :key="fIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedIndex === fIdx }"
          data-testid="field-card"
          @click="selectField(fIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium font-mono truncate">{{ fieldLabel(field) }}</div>
            </div>
            <Badge
              v-if="isComputed(field)"
              variant="secondary"
              class="shrink-0 text-[10px]"
            >
              computed
            </Badge>
            <Badge
              v-else-if="fIdx < 3"
              variant="outline"
              class="shrink-0 text-[10px]"
            >
              highlight
            </Badge>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: field detail -->
    <div class="flex-1 min-w-0">
      <div v-if="!selectedField" class="flex items-center justify-center h-full text-sm text-muted-foreground">
        Select a field to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Field Details</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete field"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-field-btn"
            @click="removeField(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input v-model="selectedField.name" placeholder="field_api_name" class="font-mono" />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Expression (CEL)</Label>
          <p class="text-[11px] text-muted-foreground">
            Data source. Can reference queries (<span class="font-mono">{{ queryNames().join(', ') || 'none' }}</span>) and other fields.
          </p>
          <ExpressionBuilder
            :model-value="selectedField.expr ?? ''"
            context="default_expr"
            height="120px"
            placeholder="main.Name + ' (' + main.Industry + ')'"
            :show-field-picker="false"
            @update:model-value="selectedField.expr = $event || undefined"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">When (CEL)</Label>
          <ExpressionBuilder
            :model-value="selectedField.when ?? ''"
            context="when_expression"
            height="80px"
            placeholder="has(main.amount)"
            :show-field-picker="false"
            @update:model-value="selectedField.when = $event || undefined"
          />
        </div>
      </div>
    </div>
  </div>
</template>
