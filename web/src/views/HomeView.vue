<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

onMounted(() => {
  const token = localStorage.getItem('crm_access_token')
  if (token) {
    router.replace('/app')
  } else {
    router.replace('/login')
  }
})
</script>

<template>
  <main class="flex items-center justify-center min-h-screen">
    <div class="text-center">
      <h1 class="text-2xl font-bold mb-2">CRM Platform</h1>
      <p class="text-muted-foreground">Loading...</p>
    </div>
  </main>
</template>
