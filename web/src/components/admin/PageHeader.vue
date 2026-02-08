<script setup lang="ts">
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'

interface BreadcrumbEntry {
  label: string
  to?: string
}

defineProps<{
  title: string
  breadcrumbs?: BreadcrumbEntry[]
}>()
</script>

<template>
  <div class="mb-6">
    <Breadcrumb v-if="breadcrumbs?.length" class="mb-2">
      <BreadcrumbList>
        <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
          <BreadcrumbItem>
            <BreadcrumbLink v-if="crumb.to" as-child>
              <RouterLink :to="crumb.to">{{ crumb.label }}</RouterLink>
            </BreadcrumbLink>
            <BreadcrumbPage v-else>{{ crumb.label }}</BreadcrumbPage>
          </BreadcrumbItem>
          <BreadcrumbSeparator v-if="idx < breadcrumbs.length - 1" />
        </template>
      </BreadcrumbList>
    </Breadcrumb>
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ title }}</h1>
      <div class="flex items-center gap-2">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>
