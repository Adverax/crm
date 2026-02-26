<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'

interface QueryResult {
  totalSize: number
  records: Record<string, unknown>[]
  error?: string
}

defineProps<{
  result: QueryResult
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div
    class="rounded-md border bg-muted/30 p-3 space-y-2"
    data-testid="soql-test-result"
  >
    <div class="flex items-center justify-between">
      <span
        v-if="!result.error"
        class="text-xs text-muted-foreground"
      >
        {{ result.totalSize }} record(s) found
      </span>
      <span
        v-else
        class="text-xs text-destructive"
      >
        Query failed
      </span>
      <IconButton
        :icon="X"
        tooltip="Close"
        variant="ghost"
        size="icon-sm"
        class="h-6 w-6"
        data-testid="close-result-btn"
        @click="emit('close')"
      />
    </div>

    <div
      v-if="result.error"
      class="text-xs text-destructive"
    >
      {{ result.error }}
    </div>

    <div
      v-else-if="result.records.length > 0"
      class="overflow-x-auto"
    >
      <table class="w-full text-xs">
        <thead>
          <tr class="border-b">
            <th
              v-for="key in Object.keys(result.records[0]!)"
              :key="key"
              class="text-left p-1 font-medium text-muted-foreground"
            >
              {{ key }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(record, idx) in result.records"
            :key="idx"
            class="border-b last:border-0"
          >
            <td
              v-for="key in Object.keys(result.records[0]!)"
              :key="key"
              class="p-1 font-mono"
            >
              {{ record[key] ?? '' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div
      v-else
      class="text-xs text-muted-foreground"
    >
      No records returned
    </div>
  </div>
</template>
