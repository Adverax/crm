<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { viewsApi } from '@/api/views'
import { useToast } from '@/composables/useToast'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import type { Portal, GateResponse } from '@/types/portals'

const props = defineProps<{
  portalApiName: string
}>()

const toast = useToast()

const view = ref<Portal | null>(null)
const gateResult = ref<GateResponse | null>(null)
const currentGate = ref<string>('')
const loading = ref(true)
const error = ref<string | null>(null)

async function loadView() {
  loading.value = true
  error.value = null
  try {
    const res = await viewsApi.getByAPIName(props.portalApiName)
    view.value = res.data
    // Execute entry gate
    const entryGate = res.data.config?.entryGate
    if (entryGate) {
      await executeGate(entryGate)
    }
  } catch (err) {
    error.value = 'Failed to load page'
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

async function executeGate(gateName: string, params?: Record<string, string>) {
  try {
    const res = await viewsApi.executeGate(props.portalApiName, gateName, params)
    gateResult.value = res.data
    currentGate.value = gateName
  } catch (err) {
    toast.errorFromApi(err)
  }
}

function onOutcomeClick(outcome: { gate?: string; name?: string }) {
  if (outcome.gate) {
    executeGate(outcome.gate)
  }
}

onMounted(loadView)
watch(() => props.portalApiName, loadView)
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

      <template v-if="gateResult">
        <!-- Datasets -->
        <div v-if="gateResult.datasets" class="space-y-4 mb-6" data-testid="page-datasets">
          <Card v-for="(dataset, name) in gateResult.datasets" :key="String(name)">
            <CardHeader>
              <CardTitle class="text-base">{{ name }}</CardTitle>
            </CardHeader>
            <CardContent>
              <pre class="text-sm text-muted-foreground font-mono whitespace-pre-wrap">{{ JSON.stringify(dataset, null, 2) }}</pre>
            </CardContent>
          </Card>
        </div>

        <!-- Outcomes -->
        <div v-if="gateResult.outcomes?.length" class="flex gap-2" data-testid="page-outcomes">
          <Button
            v-for="outcome in gateResult.outcomes"
            :key="outcome.name"
            variant="outline"
            size="sm"
            @click="onOutcomeClick(outcome)"
          >
            {{ outcome.label || outcome.name }}
          </Button>
        </div>
      </template>

      <div v-else class="text-center py-12 text-muted-foreground">
        <p>This page has no content configured yet.</p>
      </div>
    </template>
  </div>
</template>
