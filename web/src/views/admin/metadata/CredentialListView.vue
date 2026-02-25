<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { credentialsApi } from '@/api/credentials'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ActiveStatusBadge from '@/components/admin/ActiveStatusBadge.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { Credential } from '@/types/credentials'

const router = useRouter()
const toast = useToast()

const credentials = ref<Credential[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadCredentials() {
  loading.value = true
  error.value = null
  try {
    const response = await credentialsApi.list()
    credentials.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load credentials: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadCredentials)

function goToCreate() {
  router.push({ name: 'admin-credential-create' })
}

function goToDetail(id: string) {
  router.push({ name: 'admin-credential-detail', params: { credentialId: id } })
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Credentials' },
]
</script>

<template>
  <div>
    <PageHeader title="Credentials" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create credential"
          variant="default"
          size="icon-sm"
          data-testid="create-credential-btn"
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
      v-else-if="credentials.length === 0"
      title="No credentials"
      description="Create named credentials for HTTP integrations."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="cred in credentials"
        :key="cred.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="credential-row"
        @click="goToDetail(cred.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium font-mono">{{ cred.code }}</div>
            <div class="text-sm text-muted-foreground">{{ cred.name }} &mdash; {{ cred.baseUrl }}</div>
          </div>
          <div class="flex items-center gap-2">
            <Badge variant="outline">{{ cred.type }}</Badge>
            <ActiveStatusBadge :is-active="cred.isActive" />
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
