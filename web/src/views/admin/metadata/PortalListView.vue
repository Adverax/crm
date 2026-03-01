<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { portalsApi } from '@/api/portals'
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
import type { Portal } from '@/types/portals'

const router = useRouter()
const toast = useToast()

const views = ref<Portal[]>([])
const searchQuery = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const filteredViews = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return views.value
  return views.value.filter(
    (view) =>
      view.label.toLowerCase().includes(q) ||
      view.apiName.toLowerCase().includes(q),
  )
})

async function loadViews() {
  loading.value = true
  error.value = null
  try {
    const response = await portalsApi.list()
    views.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load object views: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadViews)

function goToCreate() {
  router.push({ name: 'admin-portal-create' })
}

function goToDetail(portalId: string) {
  router.push({ name: 'admin-portal-detail', params: { portalId } })
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Portals' },
]
</script>

<template>
  <div>
    <PageHeader title="Portals" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create object view"
          variant="default"
          size="icon-sm"
          data-testid="create-view-btn"
          @click="goToCreate"
        />
      </template>
    </PageHeader>

    <div class="mb-4">
      <Input
        v-model="searchQuery"
        placeholder="Filter..."
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
      v-else-if="filteredViews.length === 0"
      title="No object views"
      description="Create your first object view to customize how records are displayed per profile."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="view in filteredViews"
        :key="view.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="view-row"
        @click="goToDetail(view.id!)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium">{{ view.label }}</div>
            <div class="text-sm text-muted-foreground font-mono">
              {{ view.apiName }}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Badge v-if="view.profileId" variant="secondary">Profile-specific</Badge>
            <Badge v-else variant="outline">Global</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
