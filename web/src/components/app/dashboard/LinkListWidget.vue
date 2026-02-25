<script setup lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ResolvedWidget } from '@/types/dashboard'

defineProps<{
  widget: ResolvedWidget
}>()
</script>

<template>
  <Card data-testid="link-list-widget">
    <CardHeader class="pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground">
        {{ widget.label }}
      </CardTitle>
    </CardHeader>
    <CardContent class="p-0">
      <div v-if="!widget.links || widget.links.length === 0" class="p-4 text-sm text-muted-foreground text-center">
        No links
      </div>
      <ul v-else>
        <li v-for="link in widget.links" :key="link.url" class="border-b last:border-0">
          <RouterLink
            :to="link.url"
            class="flex items-center gap-2 px-4 py-3 text-sm hover:bg-muted/50 transition-colors"
          >
            {{ link.label }}
          </RouterLink>
        </li>
      </ul>
    </CardContent>
  </Card>
</template>
