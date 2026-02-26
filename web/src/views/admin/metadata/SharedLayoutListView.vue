<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { sharedLayoutsApi } from '@/api/shared-layouts'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import { Input } from '@/components/ui/input'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { SharedLayout } from '@/types/layouts'

const router = useRouter()
const toast = useToast()

const sharedLayouts = ref<SharedLayout[]>([])
const searchQuery = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const filteredLayouts = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return sharedLayouts.value
  return sharedLayouts.value.filter(
    (sl) =>
      sl.apiName.toLowerCase().includes(q) ||
      sl.label.toLowerCase().includes(q),
  )
})

async function loadSharedLayouts() {
  loading.value = true
  error.value = null
  try {
    const response = await sharedLayoutsApi.list()
    sharedLayouts.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load shared layouts: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadSharedLayouts)

function goToCreate() {
  router.push({ name: 'admin-shared-layout-create' })
}

function goToDetail(id: string) {
  router.push({ name: 'admin-shared-layout-detail', params: { sharedLayoutId: id } })
}

function typeVariant(type: string): 'default' | 'secondary' | 'outline' {
  if (type === 'field') return 'default'
  if (type === 'section') return 'secondary'
  return 'outline'
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Shared Layouts' },
]
</script>

<template>
  <div>
    <PageHeader title="Shared Layouts" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create shared layout"
          variant="default"
          size="icon-sm"
          data-testid="create-shared-layout-btn"
          @click="goToCreate"
        />
      </template>
    </PageHeader>

    <div class="mb-4">
      <Input
        v-model="searchQuery"
        placeholder="Filter by api_name or label..."
        class="h-9 w-64"
        data-testid="search-input"
      />
    </div>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="filteredLayouts.length === 0"
      title="No shared layouts"
      description="Create your first shared layout to reuse field, section, or list configurations across multiple layouts."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="sl in filteredLayouts"
        :key="sl.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="shared-layout-row"
        @click="goToDetail(sl.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium">{{ sl.label }}</div>
            <div class="text-sm text-muted-foreground font-mono">
              {{ sl.apiName }}
            </div>
          </div>
          <Badge :variant="typeVariant(sl.type)">
            {{ sl.type }}
          </Badge>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
