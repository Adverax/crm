<script setup lang="ts">
import { Badge } from '@/components/ui/badge'

defineProps<{
  isActive: boolean
}>()
</script>

<template>
  <Badge :variant="isActive ? 'default' : 'secondary'">
    {{ isActive ? 'Active' : 'Inactive' }}
  </Badge>
</template>
