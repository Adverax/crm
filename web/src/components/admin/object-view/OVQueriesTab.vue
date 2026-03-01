<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import SoqlEditor from '@/components/admin/soql-editor/SoqlEditor.vue'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import { ref, computed } from 'vue'
import type { OVQuery } from '@/types/object-views'

const props = defineProps<{
  queries: OVQuery[]
}>()

const emit = defineEmits<{
  'update:queries': [value: OVQuery[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedQuery = computed(() =>
  selectedIndex.value !== null ? props.queries[selectedIndex.value] ?? null : null,
)

function queryType(query: OVQuery): string {
  return /\bSELECT\s+ROW\b/i.test(query.soql) ? 'scalar' : 'list'
}

function selectQuery(index: number) {
  selectedIndex.value = index
}

function addQuery() {
  const updated: OVQuery[] = [...props.queries, { name: '', soql: '', when: '' }]
  emit('update:queries', updated)
  selectedIndex.value = updated.length - 1
}

function removeQuery(index: number) {
  const updated = [...props.queries]
  updated.splice(index, 1)
  emit('update:queries', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}
</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="queries-master-detail">
    <!-- Left panel: query list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Queries</span>
        <IconButton
          :icon="Plus"
          tooltip="Add query"
          variant="outline"
          size="sm"
          data-testid="add-query-btn"
          @click="addQuery"
        />
      </div>
      <div v-if="queries.length === 0" class="p-3 text-sm text-muted-foreground">
        No queries configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(query, qIdx) in queries"
          :key="qIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedIndex === qIdx }"
          data-testid="query-card"
          @click="selectQuery(qIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium font-mono truncate">{{ query.name || 'unnamed' }}</div>
            </div>
            <Badge
              :variant="queryType(query) === 'scalar' ? 'default' : 'secondary'"
              class="shrink-0 text-[10px]"
            >
              {{ queryType(query) }}
            </Badge>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: query detail -->
    <div class="flex-1 min-w-0">
      <div v-if="!selectedQuery" class="flex items-center justify-center h-full text-sm text-muted-foreground">
        Select a query to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Query Details</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete query"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-query-btn"
            @click="removeQuery(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input v-model="selectedQuery.name" placeholder="recent_activities" class="font-mono" />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">When (CEL)</Label>
          <ExpressionBuilder
            :model-value="selectedQuery.when ?? ''"
            context="when_expression"
            height="80px"
            placeholder="record.status == 'active'"
            @update:model-value="selectedQuery.when = $event"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">SOQL</Label>
          <SoqlEditor
            v-model="selectedQuery.soql"
            height="200px"
          />
        </div>
      </div>
    </div>
  </div>
</template>
