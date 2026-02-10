<script setup lang="ts">
import { useId } from 'vue'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import type { BitmaskFlag } from '@/types/security'

const props = defineProps<{
  flags: BitmaskFlag[]
  modelValue: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const uid = useId()

function isChecked(bit: number): boolean {
  return (props.modelValue & bit) !== 0
}

function toggle(bit: number) {
  emit('update:modelValue', props.modelValue ^ bit)
}
</script>

<template>
  <div class="flex items-center gap-4">
    <div v-for="flag in flags" :key="flag.key" class="flex items-center gap-2">
      <Switch
        :id="`${uid}-${flag.key}`"
        :checked="isChecked(flag.bit)"
        @update:checked="toggle(flag.bit)"
      />
      <Label :for="`${uid}-${flag.key}`" class="cursor-pointer">{{ flag.label }}</Label>
    </div>
  </div>
</template>
