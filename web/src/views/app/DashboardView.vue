<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { dashboardApi } from '@/api/dashboard'
import { useToast } from '@/composables/useToast'
import { Skeleton } from '@/components/ui/skeleton'
import MetricWidget from '@/components/app/dashboard/MetricWidget.vue'
import ListWidget from '@/components/app/dashboard/ListWidget.vue'
import LinkListWidget from '@/components/app/dashboard/LinkListWidget.vue'
import type { ResolvedWidget } from '@/types/dashboard'

const toast = useToast()

const widgets = ref<ResolvedWidget[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await dashboardApi.resolve()
    widgets.value = res.data?.widgets ?? []
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
})

function sizeClass(size: string): string {
  switch (size) {
    case 'full':
      return 'col-span-12'
    case 'half':
      return 'col-span-12 md:col-span-6'
    case 'third':
      return 'col-span-12 md:col-span-4'
    default:
      return 'col-span-12 md:col-span-6'
  }
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-2xl font-semibold mb-6" data-testid="dashboard-title">Dashboard</h1>

    <div v-if="loading" class="grid grid-cols-12 gap-4">
      <Skeleton class="col-span-12 md:col-span-6 h-48" />
      <Skeleton class="col-span-12 md:col-span-6 h-48" />
    </div>

    <div v-else-if="widgets.length === 0" class="text-center py-20 text-muted-foreground" data-testid="dashboard-empty">
      <p class="text-lg font-medium">Welcome to CRM</p>
      <p class="text-sm mt-1">Your dashboard will appear here when configured by an administrator.</p>
    </div>

    <div v-else class="grid grid-cols-12 gap-4" data-testid="dashboard-grid">
      <div
        v-for="widget in widgets"
        :key="widget.key"
        :class="sizeClass(widget.size)"
      >
        <MetricWidget v-if="widget.type === 'metric'" :widget="widget" />
        <ListWidget v-else-if="widget.type === 'list'" :widget="widget" />
        <LinkListWidget v-else-if="widget.type === 'link_list'" :widget="widget" />
      </div>
    </div>
  </div>
</template>
