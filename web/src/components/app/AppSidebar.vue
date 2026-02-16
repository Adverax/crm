<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { Separator } from '@/components/ui/separator'
import { Button } from '@/components/ui/button'
import { useRecordsStore } from '@/stores/records'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const recordsStore = useRecordsStore()
const authStore = useAuthStore()
const toast = useToast()
const { navObjects } = storeToRefs(recordsStore)

onMounted(async () => {
  try {
    await recordsStore.fetchNavObjects()
  } catch (err) {
    toast.errorFromApi(err)
  }
})

function isActive(apiName: string): boolean {
  return route.path.startsWith(`/app/${apiName}`)
}

async function onLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="w-60 border-r bg-muted/30 flex flex-col">
    <div class="p-4">
      <RouterLink to="/app" class="text-lg font-semibold tracking-tight">
        CRM
      </RouterLink>
    </div>
    <Separator />
    <nav class="flex-1 p-2">
      <ul class="space-y-1">
        <li v-for="obj in navObjects" :key="obj.apiName">
          <RouterLink
            :to="`/app/${obj.apiName}`"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'bg-accent text-accent-foreground': isActive(obj.apiName) }"
          >
            {{ obj.pluralLabel }}
          </RouterLink>
        </li>
      </ul>
    </nav>
    <Separator />
    <div class="p-2">
      <RouterLink
        to="/admin"
        class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
      >
        Settings
      </RouterLink>
    </div>
    <Separator />
    <div class="p-3">
      <div class="text-xs text-muted-foreground truncate mb-2">
        {{ authStore.displayName || 'User' }}
      </div>
      <Button variant="outline" size="sm" class="w-full" @click="onLogout">
        Sign out
      </Button>
    </div>
  </aside>
</template>
