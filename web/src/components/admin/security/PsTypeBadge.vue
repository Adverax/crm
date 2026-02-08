<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import type { PsType } from '@/types/security'

defineProps<{
  type: PsType
}>()

const labels: Record<PsType, string> = {
  grant: 'Grant',
  deny: 'Deny',
}
</script>

<template>
  <Badge :variant="type === 'grant' ? 'default' : 'destructive'">
    {{ labels[type] }}
  </Badge>
</template>
