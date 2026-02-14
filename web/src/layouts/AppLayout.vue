<script setup lang="ts">
import AppSidebar from '@/components/app/AppSidebar.vue'
</script>

<template>
  <div class="flex min-h-screen">
    <AppSidebar />
    <main class="flex-1 p-6 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
