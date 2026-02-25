<script setup lang="ts">
import AdminSidebar from '@/components/admin/AdminSidebar.vue'
</script>

<template>
  <div class="flex min-h-screen">
    <AdminSidebar />
    <main class="flex-1 p-6 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
:deep(input::placeholder),
:deep(textarea::placeholder) {
  color: hsl(0 0% 75%);
}
</style>
