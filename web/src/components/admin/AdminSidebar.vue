<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { reactive, watchEffect } from 'vue'
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

const groups: NavGroup[] = [
  {
    label: 'Schema',
    children: [
      { label: 'Objects', to: '/admin/metadata/objects' },
      { label: 'Functions', to: '/admin/metadata/functions' },
    ],
  },
  {
    label: 'Presentation',
    children: [
      { label: 'Object Views', to: '/admin/metadata/object-views' },
      { label: 'Layouts', to: '/admin/metadata/layouts' },
      { label: 'Shared Layouts', to: '/admin/metadata/shared-layouts' },
      { label: 'Navigation', to: '/admin/metadata/navigation' },
    ],
  },
  {
    label: 'Automation',
    children: [
      { label: 'Procedures', to: '/admin/metadata/procedures' },
      { label: 'Automation Rules', to: '/admin/metadata/automation-rules' },
      { label: 'Credentials', to: '/admin/metadata/credentials' },
    ],
  },
  {
    label: 'Security',
    children: [
      { label: 'Roles', to: '/admin/security/roles' },
      { label: 'Permission Sets', to: '/admin/security/permission-sets' },
      { label: 'Profiles', to: '/admin/security/profiles' },
      { label: 'Groups', to: '/admin/security/groups' },
      { label: 'Sharing Rules', to: '/admin/security/sharing-rules' },
    ],
  },
  {
    label: 'Territories',
    children: [
      { label: 'Models', to: '/admin/territory/models' },
      { label: 'Territories', to: '/admin/territory/territories' },
    ],
  },
]

const flatItems: NavItem[] = [
  { label: 'Templates', to: '/admin/templates' },
]

const bottomItems: NavItem[] = [
  { label: 'Users', to: '/admin/security/users' },
]

const expanded = reactive<Record<string, boolean>>({})

function isGroupActive(group: NavGroup): boolean {
  return group.children.some((child) => route.path.startsWith(child.to))
}

watchEffect(() => {
  for (const group of groups) {
    if (isGroupActive(group)) {
      expanded[group.label] = true
    }
  }
})

function isActive(path: string): boolean {
  return route.path.startsWith(path)
}

function toggleGroup(label: string) {
  expanded[label] = !expanded[label]
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
        <li v-for="group in groups" :key="group.label">
          <button
            class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'text-accent-foreground': isGroupActive(group) }"
            @click="toggleGroup(group.label)"
          >
            {{ group.label }}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-4 w-4 transition-transform"
              :class="{ 'rotate-180': expanded[group.label] }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          <ul v-if="expanded[group.label]" class="ml-3 space-y-1 mt-1">
            <li v-for="child in group.children" :key="child.to">
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

        <li v-for="item in flatItems" :key="item.to">
          <RouterLink
            :to="item.to"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'bg-accent text-accent-foreground': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
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
        {{ authStore.displayName || 'User' }}
      </div>
      <Button variant="outline" size="sm" class="w-full" @click="onLogout">
        Sign out
      </Button>
    </div>
  </aside>
</template>
