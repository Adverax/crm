<script setup lang="ts">
import { computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ResolvedWidget, MetricWidgetData } from '@/types/dashboard'

const props = defineProps<{
  widget: ResolvedWidget
}>()

const value = computed(() => {
  const data = props.widget.data as MetricWidgetData | null
  if (!data) return '—'
  const raw = data.value
  if (raw === null || raw === undefined) return '—'
  if (props.widget.format === 'percent') return `${raw}%`
  if (props.widget.format === 'currency') return `$${Number(raw).toLocaleString()}`
  return Number(raw).toLocaleString()
})
</script>

<template>
  <Card data-testid="metric-widget">
    <CardHeader class="pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground">
        {{ widget.label }}
      </CardTitle>
    </CardHeader>
    <CardContent>
      <div class="text-3xl font-bold">{{ value }}</div>
    </CardContent>
  </Card>
</template>
