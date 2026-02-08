<script setup lang="ts">
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'

interface FlagItem {
  key: string
  label: string
}

interface FlagGroup {
  title: string
  items: FlagItem[]
}

defineProps<{
  groups: FlagGroup[]
  modelValue: Record<string, boolean>
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, boolean>]
}>()

function toggle(key: string, value: boolean, current: Record<string, boolean>) {
  emit('update:modelValue', { ...current, [key]: value })
}
</script>

<template>
  <div class="space-y-6">
    <div v-for="group in groups" :key="group.title">
      <h3 class="text-sm font-medium text-muted-foreground mb-3">{{ group.title }}</h3>
      <div class="grid grid-cols-2 gap-4">
        <div
          v-for="item in group.items"
          :key="item.key"
          class="flex items-center justify-between space-x-2"
        >
          <Label :for="item.key" class="text-sm">{{ item.label }}</Label>
          <Switch
            :id="item.key"
            :checked="modelValue[item.key]"
            :disabled="disabled"
            @update:checked="toggle(item.key, $event, modelValue)"
          />
        </div>
      </div>
    </div>
  </div>
</template>
