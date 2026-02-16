<script setup lang="ts">
import type { ButtonVariants } from '@/components/ui/button'
import type { Component } from 'vue'
import { useAttrs } from 'vue'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

defineOptions({ inheritAttrs: false })

interface Props {
  icon: Component
  tooltip: string
  variant?: ButtonVariants['variant']
  size?: ButtonVariants['size']
  disabled?: boolean
  class?: string
}

withDefaults(defineProps<Props>(), {
  variant: 'ghost',
  size: 'icon-sm',
  disabled: false,
  class: '',
})

defineEmits<{
  click: [event: MouseEvent]
}>()

const attrs = useAttrs()
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Button
        v-bind="attrs"
        :variant="variant"
        :size="size"
        :disabled="disabled"
        :class="$props.class"
        @click="$emit('click', $event)"
      >
        <component :is="icon" />
        <span class="sr-only">{{ tooltip }}</span>
      </Button>
    </TooltipTrigger>
    <TooltipContent>{{ tooltip }}</TooltipContent>
  </Tooltip>
</template>
