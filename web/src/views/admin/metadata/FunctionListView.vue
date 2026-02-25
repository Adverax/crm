<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { functionsApi } from '@/api/functions'
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
import type { Function } from '@/types/functions'

const router = useRouter()
const toast = useToast()

const functions = ref<Function[]>([])
const searchQuery = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const filteredFunctions = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return functions.value
  return functions.value.filter(
    (fn) =>
      (fn.name?.toLowerCase().includes(q) ?? false) ||
      (fn.description?.toLowerCase().includes(q) ?? false),
  )
})

async function loadFunctions() {
  loading.value = true
  error.value = null
  try {
    const response = await functionsApi.list()
    functions.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load functions: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadFunctions)

function goToCreate() {
  router.push({ name: 'admin-function-create' })
}

function goToDetail(functionId: string) {
  router.push({ name: 'admin-function-detail', params: { functionId } })
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Functions' },
]
</script>

<template>
  <div>
    <PageHeader title="Functions" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create function"
          variant="default"
          size="icon-sm"
          data-testid="create-function-btn"
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
      v-else-if="filteredFunctions.length === 0"
      title="No functions"
      description="Create your first custom function for use in CEL expressions."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="fn in filteredFunctions"
        :key="fn.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="function-row"
        @click="goToDetail(fn.id!)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium font-mono">fn.{{ fn.name }}</div>
            <div v-if="fn.description" class="text-sm text-muted-foreground">
              {{ fn.description }}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Badge variant="secondary">
              {{ fn.returnType ?? 'any' }}
            </Badge>
            <Badge v-if="fn.params && fn.params.length > 0" variant="outline">
              {{ fn.params.length }} params
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
