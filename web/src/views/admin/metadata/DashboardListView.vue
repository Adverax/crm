<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { dashboardApi } from '@/api/dashboard'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import type { ProfileDashboard } from '@/types/dashboard'

const router = useRouter()
const toast = useToast()

const items = ref<ProfileDashboard[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const response = await dashboardApi.list()
    items.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load dashboard configs: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function goToCreate() {
  router.push({ name: 'admin-dashboard-create' })
}

function goToDetail(id: string) {
  router.push({ name: 'admin-dashboard-detail', params: { dashboardId: id } })
}

function widgetCount(dash: ProfileDashboard): number {
  return dash.config?.widgets?.length ?? 0
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Dashboards' },
]
</script>

<template>
  <div>
    <PageHeader title="Profile Dashboards" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create dashboard"
          variant="default"
          size="icon-sm"
          data-testid="create-dash-btn"
          @click="goToCreate"
        />
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      title="No dashboard configs"
      description="Create profile-specific dashboards with widgets for the home page."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="item in items"
        :key="item.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="dash-row"
        @click="goToDetail(item.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium font-mono">{{ item.profileId }}</div>
            <div class="text-sm text-muted-foreground">
              {{ widgetCount(item) }} widget(s)
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
