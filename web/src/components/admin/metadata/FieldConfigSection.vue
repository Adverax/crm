<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ConfigFieldDef } from '@/types/field-types'
import type { FieldConfig, ObjectDefinition } from '@/types/metadata'

const props = defineProps<{
  configFields: ConfigFieldDef[]
  modelValue: FieldConfig
  objects?: ObjectDefinition[]
  showReferencedObject?: boolean
  referencedObjectId?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: FieldConfig]
  'update:referencedObjectId': [value: string]
}>()

function updateConfig(key: string, value: unknown) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function getConfigValue(key: string): unknown {
  return (props.modelValue as Record<string, unknown>)[key]
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="showReferencedObject && objects" class="space-y-2">
      <Label>Referenced Object</Label>
      <Select
        :model-value="referencedObjectId ?? ''"
        @update:model-value="emit('update:referencedObjectId', String($event))"
      >
        <SelectTrigger>
          <SelectValue placeholder="Select object" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="obj in objects" :key="obj.id" :value="obj.id">
            {{ obj.label }} ({{ obj.apiName }})
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <template v-for="field in configFields" :key="field.key">
      <div v-if="field.type === 'number'" class="space-y-2">
        <Label :for="field.key">{{ field.label }}</Label>
        <Input
          :id="field.key"
          type="number"
          :model-value="(getConfigValue(field.key) as number | undefined) ?? ''"
          @update:model-value="updateConfig(field.key, $event ? Number($event) : undefined)"
        />
      </div>

      <div v-else-if="field.type === 'text'" class="space-y-2">
        <Label :for="field.key">{{ field.label }}</Label>
        <Input
          :id="field.key"
          :model-value="(getConfigValue(field.key) as string | undefined) ?? ''"
          @update:model-value="updateConfig(field.key, $event ? String($event) : undefined)"
        />
      </div>

      <div v-else-if="field.type === 'boolean'" class="flex items-center justify-between space-x-2">
        <Label :for="field.key">{{ field.label }}</Label>
        <Switch
          :id="field.key"
          :checked="(getConfigValue(field.key) as boolean) ?? false"
          @update:checked="updateConfig(field.key, $event)"
        />
      </div>

      <div v-else-if="field.type === 'select' && field.options" class="space-y-2">
        <Label :for="field.key">{{ field.label }}</Label>
        <Select
          :model-value="(getConfigValue(field.key) as string) ?? ''"
          @update:model-value="updateConfig(field.key, $event ? String($event) : undefined)"
        >
          <SelectTrigger>
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="opt in field.options" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </template>
  </div>
</template>
