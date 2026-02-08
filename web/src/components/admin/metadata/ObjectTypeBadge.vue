<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import type { ObjectType } from '@/types/metadata'

defineProps<{
  type: ObjectType
}>()

const labels: Record<ObjectType, string> = {
  standard: 'Standard',
  custom: 'Custom',
}
</script>

<template>
  <Badge :variant="type === 'standard' ? 'default' : 'secondary'">
    {{ labels[type] }}
  </Badge>
</template>
