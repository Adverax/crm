<script setup lang="ts">
import type { CelValidateError } from '@/types/functions'

defineProps<{
  errors: CelValidateError[]
}>()
</script>

<template>
  <div
    v-if="errors.length > 0"
    class="rounded-md border border-destructive/50 bg-destructive/10 px-3 py-2 space-y-1"
    data-testid="expression-errors"
  >
    <div
      v-for="(err, idx) in errors"
      :key="idx"
      class="text-xs text-destructive"
    >
      <span v-if="err.line != null" class="font-mono">
        Строка {{ err.line }}<span v-if="err.column != null">:{{ err.column }}</span> —
      </span>
      {{ err.message }}
    </div>
  </div>
</template>
