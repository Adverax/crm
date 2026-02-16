<script setup lang="ts">
import { Badge } from '@/components/ui/badge'

defineProps<{
  active: boolean
}>()
</script>

<template>
  <Badge :variant="active ? 'default' : 'secondary'">
    {{ active ? 'Active' : 'Inactive' }}
  </Badge>
</template>
