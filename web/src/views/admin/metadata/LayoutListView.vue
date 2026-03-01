<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { layoutsApi } from '@/api/layouts'
import { objectViewsApi } from '@/api/object-views'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { Layout } from '@/types/layouts'
import type { ObjectView } from '@/types/object-views'

const router = useRouter()
const route = useRoute()
const toast = useToast()

const layouts = ref<Layout[]>([])
const objectViews = ref<ObjectView[]>([])
const searchQuery = ref('')
const filterOvId = ref('__all__')
const loading = ref(false)
const error = ref<string | null>(null)

const filteredLayouts = computed(() => {
  let result = layouts.value

  if (filterOvId.value && filterOvId.value !== '__all__') {
    result = result.filter((l) => l.objectViewId === filterOvId.value)
  }

  const q = searchQuery.value.toLowerCase()
  if (q) {
    result = result.filter((l) => {
      const ovLabel = ovLabelMap.value[l.objectViewId] ?? ''
      return (
        l.formFactor.toLowerCase().includes(q) ||
        l.mode.toLowerCase().includes(q) ||
        ovLabel.toLowerCase().includes(q)
      )
    })
  }

  return result
})

const ovLabelMap = computed(() => {
  const map: Record<string, string> = {}
  for (const ov of objectViews.value) {
    if (ov.id) {
      map[ov.id] = ov.label
    }
  }
  return map
})

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const queryOvId = route.query.object_view_id as string | undefined
    if (queryOvId) {
      filterOvId.value = queryOvId
    }

    const [layoutsRes, ovsRes] = await Promise.all([
      layoutsApi.list(queryOvId),
      objectViewsApi.list(),
    ])
    layouts.value = layoutsRes.data ?? []
    objectViews.value = ovsRes.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load layouts: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function goToCreate() {
  router.push({ name: 'admin-layout-create' })
}

function goToDetail(layoutId: string) {
  router.push({ name: 'admin-layout-detail', params: { layoutId } })
}

function formFactorVariant(ff: string): 'default' | 'secondary' | 'outline' {
  if (ff === 'desktop') return 'default'
  if (ff === 'tablet') return 'secondary'
  return 'outline'
}

function modeVariant(mode: string): 'default' | 'secondary' | 'outline' {
  if (mode === 'read') return 'default'
  return 'outline'
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onOvFilterChange(value: any) {
  filterOvId.value = String(value) || '__all__'
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Layouts' },
]
</script>

<template>
  <div>
    <PageHeader title="Layouts" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create layout"
          variant="default"
          size="icon-sm"
          data-testid="create-layout-btn"
          @click="goToCreate"
        />
      </template>
    </PageHeader>

    <div class="mb-4 flex gap-3 items-center">
      <Input
        v-model="searchQuery"
        placeholder="Filter..."
        class="h-9 w-64"
        data-testid="search-input"
      />
      <Select :model-value="filterOvId" @update:model-value="onOvFilterChange">
        <SelectTrigger class="w-64 h-9" data-testid="filter-ov">
          <SelectValue placeholder="All Object Views" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">All Object Views</SelectItem>
          <SelectItem
            v-for="ov in objectViews"
            :key="ov.id"
            :value="ov.id!"
          >
            {{ ov.label }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="filteredLayouts.length === 0"
      title="No layouts"
      description="Create your first layout to customize how forms are rendered per form factor and mode."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="layout in filteredLayouts"
        :key="layout.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="layout-row"
        @click="goToDetail(layout.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium">
              {{ ovLabelMap[layout.objectViewId] || layout.objectViewId }}
            </div>
            <div class="text-sm text-muted-foreground">
              Layout for {{ layout.formFactor }} / {{ layout.mode }}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Badge :variant="formFactorVariant(layout.formFactor)">
              {{ layout.formFactor }}
            </Badge>
            <Badge :variant="modeVariant(layout.mode)">
              {{ layout.mode }}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
