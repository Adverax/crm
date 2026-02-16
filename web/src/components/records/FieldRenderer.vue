<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { FieldDescribe } from '@/types/records'

const props = defineProps<{
  field: FieldDescribe
  modelValue: unknown
}>()

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
}>()

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

function onNumberInput(event: Event) {
  const target = event.target as HTMLInputElement
  const val = target.value
  if (val === '') {
    emit('update:modelValue', null)
  } else {
    emit('update:modelValue', Number(val))
  }
}

function onSwitchChange(checked: boolean) {
  emit('update:modelValue', checked)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onSelectChange(value: any) {
  emit('update:modelValue', String(value))
}

function getInputType(): string {
  const { fieldType, fieldSubtype } = props.field
  if (fieldType === 'text') {
    if (fieldSubtype === 'email') return 'email'
    if (fieldSubtype === 'phone') return 'tel'
    if (fieldSubtype === 'url') return 'url'
    return 'text'
  }
  if (fieldType === 'datetime') {
    if (fieldSubtype === 'date') return 'date'
    return 'datetime-local'
  }
  return 'text'
}

function getNumberStep(): string {
  const { fieldSubtype } = props.field
  if (fieldSubtype === 'integer') return '1'
  return '0.01'
}

function isTextArea(): boolean {
  return props.field.fieldType === 'text' && props.field.fieldSubtype === 'area'
}

function isBoolean(): boolean {
  return props.field.fieldType === 'boolean'
}

function isNumber(): boolean {
  return props.field.fieldType === 'number'
}

function isPicklist(): boolean {
  return props.field.fieldType === 'picklist'
}

function isTextInput(): boolean {
  return (
    (props.field.fieldType === 'text' && props.field.fieldSubtype !== 'area') ||
    props.field.fieldType === 'datetime' ||
    props.field.fieldType === 'reference'
  )
}
</script>

<template>
  <div class="space-y-2">
    <Label :for="field.apiName">
      {{ field.label }}
      <span v-if="field.isRequired" class="text-destructive">*</span>
    </Label>

    <Textarea
      v-if="isTextArea()"
      :id="field.apiName"
      :model-value="String(modelValue ?? '')"
      :required="field.isRequired"
      rows="3"
      @input="onInput"
    />

    <div v-else-if="isBoolean()" class="flex items-center gap-2 pt-1">
      <Switch
        :id="field.apiName"
        :checked="Boolean(modelValue)"
        @update:checked="onSwitchChange"
      />
    </div>

    <Input
      v-else-if="isNumber()"
      :id="field.apiName"
      type="number"
      :step="getNumberStep()"
      :model-value="modelValue != null ? String(modelValue) : ''"
      :required="field.isRequired"
      @input="onNumberInput"
    />

    <Select
      v-else-if="isPicklist()"
      :model-value="modelValue != null ? String(modelValue) : undefined"
      @update:model-value="onSelectChange"
    >
      <SelectTrigger>
        <SelectValue :placeholder="`Select ${field.label.toLowerCase()}`" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem
          v-for="opt in (field.config.values ?? [])"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </SelectItem>
      </SelectContent>
    </Select>

    <Input
      v-else-if="isTextInput()"
      :id="field.apiName"
      :type="getInputType()"
      :model-value="String(modelValue ?? '')"
      :required="field.isRequired"
      @input="onInput"
    />
  </div>
</template>
