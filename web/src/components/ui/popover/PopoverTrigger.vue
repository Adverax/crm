<script setup lang="ts">
import type { PopoverTriggerProps } from "reka-ui"
import { PopoverTrigger, useForwardProps } from "reka-ui"

const props = defineProps<PopoverTriggerProps>()

const forwardedProps = useForwardProps(props)
</script>

<template>
  <PopoverTrigger class="outline-none" v-bind="forwardedProps">
    <slot />
  </PopoverTrigger>
</template>
