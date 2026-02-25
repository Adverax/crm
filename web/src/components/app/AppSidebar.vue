<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ChevronDown, ChevronRight } from 'lucide-vue-next'
import { Separator } from '@/components/ui/separator'
import { Button } from '@/components/ui/button'
import { useRecordsStore } from '@/stores/records'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { navigationApi } from '@/api/navigation'
import type { ResolvedNavGroup, ResolvedNavItem } from '@/types/navigation'

const route = useRoute()
const router = useRouter()
const recordsStore = useRecordsStore()
const authStore = useAuthStore()
const toast = useToast()
const { navObjects } = storeToRefs(recordsStore)

const navGroups = ref<ResolvedNavGroup[]>([])
const hasCustomNav = ref(false)
const collapsedGroups = ref<Set<string>>(new Set())

onMounted(async () => {
  try {
    const res = await navigationApi.resolve()
    const groups = res.data?.groups ?? []
    if (groups.length > 0 && !(groups.length === 1 && groups[0]!.label === '')) {
      navGroups.value = groups
      hasCustomNav.value = true
    } else {
      // Fallback: load flat object list
      await recordsStore.fetchNavObjects()
      hasCustomNav.value = false
    }
  } catch {
    // Fallback on error
    try {
      await recordsStore.fetchNavObjects()
    } catch (err) {
      toast.errorFromApi(err)
    }
    hasCustomNav.value = false
  }
})

function isActive(apiName: string): boolean {
  return route.path.startsWith(`/app/${apiName}`)
}

function isLinkActive(url: string): boolean {
  return route.path === url
}

function toggleGroup(key: string) {
  if (collapsedGroups.value.has(key)) {
    collapsedGroups.value.delete(key)
  } else {
    collapsedGroups.value.add(key)
  }
}

function getItemRoute(item: ResolvedNavItem): string {
  if (item.type === 'object' && item.objectApiName) {
    return `/app/${item.objectApiName}`
  }
  if (item.type === 'link' && item.url) {
    return item.url
  }
  return '#'
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
    <nav class="flex-1 p-2 overflow-y-auto">
      <!-- Custom grouped navigation -->
      <template v-if="hasCustomNav">
        <div v-for="group in navGroups" :key="group.key" class="mb-2">
          <button
            class="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors"
            @click="toggleGroup(group.key)"
          >
            <span>{{ group.label }}</span>
            <component
              :is="collapsedGroups.has(group.key) ? ChevronRight : ChevronDown"
              class="h-3.5 w-3.5"
            />
          </button>
          <ul v-show="!collapsedGroups.has(group.key)" class="space-y-0.5 mt-0.5">
            <template v-for="(item, idx) in group.items" :key="`${group.key}-${idx}`">
              <li v-if="item.type === 'divider'">
                <Separator class="my-1" />
              </li>
              <li v-else>
                <RouterLink
                  :to="getItemRoute(item)"
                  class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                  :class="{
                    'bg-accent text-accent-foreground':
                      (item.type === 'object' && item.objectApiName && isActive(item.objectApiName)) ||
                      (item.type === 'link' && item.url && isLinkActive(item.url))
                  }"
                >
                  {{ item.type === 'object' ? (item.pluralLabel || item.label) : item.label }}
                </RouterLink>
              </li>
            </template>
          </ul>
        </div>
      </template>

      <!-- Fallback flat navigation -->
      <ul v-else class="space-y-1">
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
