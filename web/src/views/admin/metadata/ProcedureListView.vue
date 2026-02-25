<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { proceduresApi } from '@/api/procedures'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { Procedure } from '@/types/procedures'

const router = useRouter()
const toast = useToast()

const procedures = ref<Procedure[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadProcedures() {
  loading.value = true
  error.value = null
  try {
    const response = await proceduresApi.list()
    procedures.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load procedures: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadProcedures)

function goToCreate() {
  router.push({ name: 'admin-procedure-create' })
}

function goToDetail(id: string) {
  router.push({ name: 'admin-procedure-detail', params: { procedureId: id } })
}

function getStatusVariant(proc: Procedure): 'default' | 'secondary' | 'outline' {
  if (proc.publishedVersionId) return 'default'
  if (proc.draftVersionId) return 'secondary'
  return 'outline'
}

function getStatusLabel(proc: Procedure): string {
  if (proc.publishedVersionId && proc.draftVersionId) return 'Published + Draft'
  if (proc.publishedVersionId) return 'Published'
  if (proc.draftVersionId) return 'Draft'
  return 'Empty'
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Procedures' },
]
</script>

<template>
  <div>
    <PageHeader title="Procedures" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create procedure"
          variant="default"
          size="icon-sm"
          data-testid="create-procedure-btn"
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
      v-else-if="procedures.length === 0"
      title="No procedures"
      description="Create your first procedure to automate business processes."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="proc in procedures"
        :key="proc.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="procedure-row"
        @click="goToDetail(proc.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium font-mono">{{ proc.code }}</div>
            <div class="text-sm text-muted-foreground">{{ proc.name }}</div>
          </div>
          <Badge :variant="getStatusVariant(proc)">
            {{ getStatusLabel(proc) }}
          </Badge>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
