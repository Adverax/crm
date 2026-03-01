<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Badge } from '@/components/ui/badge'
import { X } from 'lucide-vue-next'
import type { DmlTestResponse } from '@/api/dml'

defineProps<{
  result: DmlTestResponse
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div
    class="rounded-md border bg-muted/30 p-3 space-y-2"
    data-testid="dml-test-result"
  >
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span
          v-if="!result.error"
          class="text-xs text-muted-foreground"
        >
          {{ result.operation }} {{ result.object }} — {{ result.rowsAffected }} row(s) affected
        </span>
        <span
          v-else
          class="text-xs text-destructive"
        >
          Execution failed
        </span>
        <Badge v-if="result.rolledBack" variant="outline" class="text-[10px]">
          Rolled back
        </Badge>
      </div>
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
  </div>
</template>
