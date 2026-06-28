<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import SoqlEditor from '@/components/admin/soql-editor/SoqlEditor.vue'
import DmlEditor from '@/components/admin/dml-editor/DmlEditor.vue'

const props = withDefaults(
  defineProps<{
    modelValue: string
    detectedType: 'soql' | 'dml' | null
    height?: string
    disabled?: boolean
  }>(),
  {
    height: '120px',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const soqlRef = ref<InstanceType<typeof SoqlEditor> | null>(null)
const dmlRef = ref<InstanceType<typeof DmlEditor> | null>(null)
const savedOffset = ref(0)

function activeEditor() {
  return props.detectedType === 'dml' ? dmlRef.value : soqlRef.value
}

function onInput(value: string) {
  // Capture cursor from current editor before type might change
  const editor = activeEditor()
  if (editor) {
    savedOffset.value = editor.getCursorOffset()
  }
  emit('update:modelValue', value)
}

// When type switches, restore cursor position and focus in the new editor
watch(
  () => props.detectedType,
  () => {
    nextTick(() => {
      const editor = activeEditor()
      if (editor) {
        editor.setCursorAt(savedOffset.value)
      }
    })
  },
)
</script>

<template>
  <div data-testid="step-instruction-editor">
    <DmlEditor
      v-if="detectedType === 'dml'"
      ref="dmlRef"
      :model-value="modelValue"
      :height="height"
      :disabled="disabled"
      @update:model-value="onInput"
    />
    <SoqlEditor
      v-else
      ref="soqlRef"
      :model-value="modelValue"
      :height="height"
      :disabled="disabled"
      @update:model-value="onInput"
    />
  </div>
</template>
