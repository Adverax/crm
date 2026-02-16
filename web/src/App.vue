<script setup lang="ts">
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'
</script>

<template>
  <TooltipProvider :delay-duration="200">
    <RouterView />
    <Toaster position="top-right" :duration="4000" />
  </TooltipProvider>
</template>
