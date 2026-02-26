<script setup lang="ts">
import { computed } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { IconButton } from '@/components/ui/icon-button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Plus, X, GripVertical } from 'lucide-vue-next'
import type { ListConfig, ListColumnConfig, ListSearchConfig } from '@/types/layouts'
import type { SectionField } from './SectionCard.vue'

const props = defineProps<{
  listConfig: ListConfig
  availableFields: SectionField[]
}>()

const emit = defineEmits<{
  'update:listConfig': [config: ListConfig]
}>()

function update(patch: Partial<ListConfig>) {
  emit('update:listConfig', { ...props.listConfig, ...patch })
}

const columns = computed({
  get: () => props.listConfig.columns ?? [],
  set: (v: ListColumnConfig[]) => update({ columns: v }),
})

const activeFieldNames = computed(() =>
  new Set(columns.value.map((c) => c.field)),
)

const unusedFields = computed(() =>
  props.availableFields.filter((f) => !activeFieldNames.value.has(f.apiName)),
)

function addColumn(fieldName: string) {
  const field = props.availableFields.find((f) => f.apiName === fieldName)
  const col: ListColumnConfig = {
    field: fieldName,
    label: field?.label,
  }
  update({ columns: [...columns.value, col] })
}

function removeColumn(index: number) {
  const updated = [...columns.value]
  updated.splice(index, 1)
  update({ columns: updated })
}

function updateColumn(index: number, patch: Partial<ListColumnConfig>) {
  const updated = [...columns.value]
  const merged = { ...updated[index], ...patch }
  updated[index] = { field: merged.field ?? '', ...merged }
  update({ columns: updated })
}

function onDragEnd() {
  update({ columns: [...columns.value] })
}

const searchFields = computed({
  get: () => props.listConfig.search?.fields?.join(', ') ?? '',
  set: (v: string) => {
    const fields = v.split(',').map((s) => s.trim()).filter(Boolean)
    const search: ListSearchConfig | undefined = fields.length
      ? { ...props.listConfig.search, fields }
      : undefined
    update({ search })
  },
})

const searchPlaceholder = computed({
  get: () => props.listConfig.search?.placeholder ?? '',
  set: (v: string) => {
    if (!props.listConfig.search?.fields?.length) return
    update({
      search: {
        ...props.listConfig.search!,
        placeholder: v || undefined,
      },
    })
  },
})
</script>

<template>
  <div class="flex gap-4" data-testid="list-config-tab">
    <!-- Available fields (left) -->
    <Card class="w-64 shrink-0">
      <CardHeader class="py-3 px-4">
        <CardTitle class="text-sm">Available Fields</CardTitle>
      </CardHeader>
      <CardContent class="px-4 pb-3">
        <div v-if="unusedFields.length" class="space-y-1">
          <div
            v-for="field in unusedFields"
            :key="field.apiName"
            class="flex items-center justify-between px-2 py-1.5 rounded text-sm hover:bg-muted cursor-pointer"
            data-testid="available-field"
            @click="addColumn(field.apiName)"
          >
            <span class="truncate">{{ field.label }}</span>
            <Plus class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
          </div>
        </div>
        <p v-else class="text-xs text-muted-foreground italic">All fields added</p>
      </CardContent>
    </Card>

    <!-- Active columns (right) -->
    <div class="flex-1 min-w-0 space-y-4">
      <Card>
        <CardHeader class="py-3 px-4">
          <CardTitle class="text-sm">Active Columns</CardTitle>
        </CardHeader>
        <CardContent class="px-4 pb-3">
          <VueDraggable
            v-model="columns"
            :animation="150"
            handle=".drag-handle"
            class="space-y-2"
            data-testid="active-columns"
            @end="onDragEnd"
          >
            <div
              v-for="(col, idx) in columns"
              :key="col.field"
              class="border rounded-md p-3"
              data-testid="active-column"
            >
              <div class="flex items-center gap-2 mb-2">
                <GripVertical class="h-4 w-4 text-muted-foreground drag-handle cursor-grab" />
                <span class="font-medium text-sm flex-1">{{ col.label || col.field }}</span>
                <Badge variant="secondary" class="text-xs">{{ col.field }}</Badge>
                <IconButton
                  :icon="X"
                  tooltip="Remove column"
                  variant="ghost"
                  size="sm"
                  data-testid="remove-column"
                  @click="removeColumn(idx)"
                />
              </div>
              <div class="grid grid-cols-3 gap-2">
                <div>
                  <Label class="text-xs">Width</Label>
                  <Input
                    :model-value="col.width ?? ''"
                    placeholder="auto"
                    class="h-8 text-xs"
                    data-testid="column-width"
                    @update:model-value="updateColumn(idx, { width: String($event) || undefined })"
                  />
                </div>
                <div>
                  <Label class="text-xs">Align</Label>
                  <Select
                    :model-value="col.align ?? 'left'"
                    @update:model-value="updateColumn(idx, { align: String($event ?? 'left') })"
                  >
                    <SelectTrigger class="h-8 text-xs" data-testid="column-align">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="left">Left</SelectItem>
                      <SelectItem value="center">Center</SelectItem>
                      <SelectItem value="right">Right</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label class="text-xs">Label</Label>
                  <Input
                    :model-value="col.label ?? ''"
                    placeholder="auto"
                    class="h-8 text-xs"
                    data-testid="column-label"
                    @update:model-value="updateColumn(idx, { label: String($event) || undefined })"
                  />
                </div>
              </div>
              <div class="flex items-center gap-4 mt-2">
                <div class="flex items-center gap-2">
                  <Label class="text-xs">Sortable</Label>
                  <Switch
                    :checked="col.sortable ?? false"
                    @update:checked="updateColumn(idx, { sortable: $event })"
                  />
                </div>
                <div v-if="col.sortable">
                  <Select
                    :model-value="col.sortDir ?? 'asc'"
                    @update:model-value="updateColumn(idx, { sortDir: String($event ?? 'asc') })"
                  >
                    <SelectTrigger class="h-7 text-xs w-20">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="asc">Asc</SelectItem>
                      <SelectItem value="desc">Desc</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>
          </VueDraggable>
          <p v-if="!columns.length" class="text-xs text-muted-foreground italic py-2" data-testid="no-columns">
            No columns configured. Click fields on the left to add.
          </p>
        </CardContent>
      </Card>

      <!-- Search config -->
      <Card>
        <CardHeader class="py-3 px-4">
          <CardTitle class="text-sm">Search</CardTitle>
        </CardHeader>
        <CardContent class="px-4 pb-4 space-y-3">
          <div>
            <Label for="search-fields">Search fields</Label>
            <Input
              id="search-fields"
              :model-value="searchFields"
              placeholder="Name, Email (comma-separated)"
              data-testid="search-fields"
              @update:model-value="searchFields = String($event)"
            />
          </div>
          <div>
            <Label for="search-placeholder">Placeholder</Label>
            <Input
              id="search-placeholder"
              :model-value="searchPlaceholder"
              placeholder="Search..."
              data-testid="search-placeholder"
              @update:model-value="searchPlaceholder = String($event)"
            />
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
