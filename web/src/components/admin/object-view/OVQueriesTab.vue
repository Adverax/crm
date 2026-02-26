<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import SoqlEditor from '@/components/admin/soql-editor/SoqlEditor.vue'
import type { OVQuery } from '@/types/object-views'

const props = defineProps<{
  queries: OVQuery[]
}>()

const emit = defineEmits<{
  'update:queries': [value: OVQuery[]]
}>()

function addQuery() {
  const updated = [...props.queries, { name: '', soql: '', when: '' }]
  emit('update:queries', updated)
}

function removeQuery(index: number) {
  const updated = [...props.queries]
  updated.splice(index, 1)
  emit('update:queries', updated)
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Queries</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add query"
        variant="outline"
        data-testid="add-query-btn"
        @click="addQuery"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      Named SOQL queries scoped to this Object View context.
    </p>

    <Card
      v-for="(query, idx) in queries"
      :key="idx"
      data-testid="query-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Name</Label>
            <Input v-model="query.name" placeholder="recent_activities" class="font-mono" />
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">When (CEL)</Label>
              <Input v-model="query.when" placeholder="record.status == 'active'" class="font-mono" />
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete query"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeQuery(idx)"
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs">SOQL</Label>
          <SoqlEditor
            v-model="query.soql"
            height="100px"
          />
        </div>
      </CardContent>
    </Card>

    <div v-if="queries.length === 0" class="text-sm text-muted-foreground">
      No queries configured. Queries define what data this view reads.
    </div>
  </div>
</template>
