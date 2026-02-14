<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed, ref, watchEffect } from 'vue'
import { Separator } from '@/components/ui/separator'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

interface NavItem {
  label: string
  to: string
}

interface NavGroup {
  label: string
  children: NavItem[]
}

const topItems: NavItem[] = [
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: 'Шаблоны', to: '/admin/templates' },
]

const securityGroup: NavGroup = {
  label: 'Безопасность',
  children: [
    { label: 'Роли', to: '/admin/security/roles' },
    { label: 'Наборы разрешений', to: '/admin/security/permission-sets' },
    { label: 'Профили', to: '/admin/security/profiles' },
    { label: 'Группы', to: '/admin/security/groups' },
    { label: 'Правила доступа', to: '/admin/security/sharing-rules' },
  ],
}

const territoryGroup: NavGroup = {
  label: 'Территории',
  children: [
    { label: 'Модели', to: '/admin/territory/models' },
    { label: 'Территории', to: '/admin/territory/territories' },
  ],
}

const bottomItems: NavItem[] = [
  { label: 'Пользователи', to: '/admin/security/users' },
]

const securityExpanded = ref(false)
const territoryExpanded = ref(false)

const isSecurityActive = computed(() =>
  securityGroup.children.some((child) => route.path.startsWith(child.to)),
)

const isTerritoryActive = computed(() =>
  territoryGroup.children.some((child) => route.path.startsWith(child.to)),
)

watchEffect(() => {
  if (isSecurityActive.value) {
    securityExpanded.value = true
  }
  if (isTerritoryActive.value) {
    territoryExpanded.value = true
  }
})

function isActive(path: string): boolean {
  return route.path.startsWith(path)
}

function toggleSecurity() {
  securityExpanded.value = !securityExpanded.value
}

function toggleTerritory() {
  territoryExpanded.value = !territoryExpanded.value
}

async function onLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="w-60 border-r bg-muted/30 flex flex-col">
    <div class="p-4">
      <RouterLink to="/admin" class="text-lg font-semibold tracking-tight">
        CRM Admin
      </RouterLink>
    </div>
    <Separator />
    <div class="p-2">
      <RouterLink
        to="/app"
        class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
      >
        CRM
      </RouterLink>
    </div>
    <Separator />
    <nav class="flex-1 p-2">
      <ul class="space-y-1">
        <li v-for="item in topItems" :key="item.to">
          <RouterLink
            :to="item.to"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'bg-accent text-accent-foreground': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
        </li>

        <li>
          <button
            class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'text-accent-foreground': isSecurityActive }"
            @click="toggleSecurity"
          >
            {{ securityGroup.label }}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-4 w-4 transition-transform"
              :class="{ 'rotate-180': securityExpanded }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          <ul v-if="securityExpanded" class="ml-3 space-y-1 mt-1">
            <li v-for="child in securityGroup.children" :key="child.to">
              <RouterLink
                :to="child.to"
                class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                :class="{ 'bg-accent text-accent-foreground': isActive(child.to) }"
              >
                {{ child.label }}
              </RouterLink>
            </li>
          </ul>
        </li>

        <li>
          <button
            class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'text-accent-foreground': isTerritoryActive }"
            @click="toggleTerritory"
          >
            {{ territoryGroup.label }}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-4 w-4 transition-transform"
              :class="{ 'rotate-180': territoryExpanded }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          <ul v-if="territoryExpanded" class="ml-3 space-y-1 mt-1">
            <li v-for="child in territoryGroup.children" :key="child.to">
              <RouterLink
                :to="child.to"
                class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                :class="{ 'bg-accent text-accent-foreground': isActive(child.to) }"
              >
                {{ child.label }}
              </RouterLink>
            </li>
          </ul>
        </li>

        <li v-for="item in bottomItems" :key="item.to">
          <RouterLink
            :to="item.to"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'bg-accent text-accent-foreground': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
        </li>
      </ul>
    </nav>
    <Separator />
    <div class="p-3">
      <div class="text-xs text-muted-foreground truncate mb-2">
        {{ authStore.displayName || 'Пользователь' }}
      </div>
      <Button variant="outline" size="sm" class="w-full" @click="onLogout">
        Выйти
      </Button>
    </div>
  </aside>
</template>
