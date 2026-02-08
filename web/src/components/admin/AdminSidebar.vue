<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import { Separator } from '@/components/ui/separator'

const route = useRoute()

const items = [
  { label: 'Объекты', to: '/admin/metadata/objects', active: true },
  { label: 'Безопасность', to: '', disabled: true },
  { label: 'Пользователи', to: '', disabled: true },
]

const isActive = computed(() => (path: string) =>
  path && route.path.startsWith(path),
)
</script>

<template>
  <aside class="w-60 border-r bg-muted/30 flex flex-col">
    <div class="p-4">
      <RouterLink to="/admin" class="text-lg font-semibold tracking-tight">
        CRM Admin
      </RouterLink>
    </div>
    <Separator />
    <nav class="flex-1 p-2">
      <ul class="space-y-1">
        <li v-for="item in items" :key="item.label">
          <RouterLink
            v-if="!item.disabled"
            :to="item.to"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="{ 'bg-accent text-accent-foreground': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
          <span
            v-else
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground cursor-not-allowed"
          >
            {{ item.label }}
          </span>
        </li>
      </ul>
    </nav>
  </aside>
</template>
