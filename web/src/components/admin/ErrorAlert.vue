<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div class="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
    {{ message }}
  </div>
</template>
