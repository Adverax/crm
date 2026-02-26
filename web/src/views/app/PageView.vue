<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { viewsApi } from '@/api/views'
import { useToast } from '@/composables/useToast'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ObjectView } from '@/types/object-views'

const props = defineProps<{
  ovApiName: string
}>()

const toast = useToast()

const view = ref<ObjectView | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

async function loadView() {
  loading.value = true
  error.value = null
  try {
    const res = await viewsApi.getByAPIName(props.ovApiName)
    view.value = res.data
  } catch (err) {
    error.value = 'Failed to load page'
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadView)
watch(() => props.ovApiName, loadView)
</script>

<template>
  <div class="p-6">
    <div v-if="loading" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-48 w-full" />
    </div>

    <div v-else-if="error" class="text-center py-20 text-muted-foreground">
      <p class="text-lg font-medium">Page not found</p>
      <p class="text-sm mt-1">The requested page could not be loaded.</p>
    </div>

    <template v-else-if="view">
      <h1 class="text-2xl font-semibold mb-6" data-testid="page-title">{{ view.label }}</h1>

      <div v-if="view.description" class="text-muted-foreground mb-6">
        {{ view.description }}
      </div>

      <div v-if="view.config?.read?.queries?.length" class="space-y-4" data-testid="page-queries">
        <Card v-for="(query, idx) in view.config.read.queries" :key="idx">
          <CardHeader>
            <CardTitle class="text-base">{{ query.name || `Query ${idx + 1}` }}</CardTitle>
          </CardHeader>
          <CardContent>
            <p class="text-sm text-muted-foreground font-mono">{{ query.soql }}</p>
          </CardContent>
        </Card>
      </div>

      <div v-else class="text-center py-12 text-muted-foreground">
        <p>This page has no content configured yet.</p>
      </div>
    </template>
  </div>
</template>
