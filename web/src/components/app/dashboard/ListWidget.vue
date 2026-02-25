<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ResolvedWidget, ListWidgetData } from '@/types/dashboard'

const props = defineProps<{
  widget: ResolvedWidget
}>()

const router = useRouter()

const data = computed(() => {
  return (props.widget.data as ListWidgetData | null) ?? { records: [], totalCount: 0 }
})

const columns = computed(() => props.widget.columns ?? [])

function goToRecord(record: Record<string, unknown>) {
  if (props.widget.objectApiName && record.id) {
    router.push(`/app/${props.widget.objectApiName}/${record.id}`)
  }
}

function cellValue(record: Record<string, unknown>, col: string): string {
  const val = record[col]
  if (val === null || val === undefined) return '—'
  return String(val)
}
</script>

<template>
  <Card data-testid="list-widget">
    <CardHeader class="pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground flex items-center justify-between">
        <span>{{ widget.label }}</span>
        <span class="text-xs font-normal">{{ data.totalCount }} total</span>
      </CardTitle>
    </CardHeader>
    <CardContent class="p-0">
      <div v-if="data.records.length === 0" class="p-4 text-sm text-muted-foreground text-center">
        No records
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b">
            <th
              v-for="col in columns"
              :key="col"
              class="text-left px-4 py-2 font-medium text-muted-foreground"
            >
              {{ col }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(record, idx) in data.records"
            :key="idx"
            class="border-b last:border-0 hover:bg-muted/50 cursor-pointer transition-colors"
            @click="goToRecord(record)"
          >
            <td v-for="col in columns" :key="col" class="px-4 py-2">
              {{ cellValue(record, col) }}
            </td>
          </tr>
        </tbody>
      </table>
    </CardContent>
  </Card>
</template>
