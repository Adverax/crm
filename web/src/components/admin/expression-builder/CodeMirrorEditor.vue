<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState, type Extension } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { bracketMatching } from '@codemirror/language'
import { celLanguage } from '@/lib/codemirror/cel-language'

const props = withDefaults(
  defineProps<{
    modelValue: string
    extensions?: Extension[]
    height?: string
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    extensions: () => [],
    height: '120px',
    placeholder: '',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorContainer = ref<HTMLDivElement | null>(null)
const view = shallowRef<EditorView | null>(null)
let isUpdating = false

function createExtensions(): Extension[] {
  return [
    celLanguage,
    history(),
    bracketMatching(),
    keymap.of([...defaultKeymap, ...historyKeymap]),
    EditorView.lineWrapping,
    EditorView.updateListener.of((update) => {
      if (update.docChanged && !isUpdating) {
        emit('update:modelValue', update.state.doc.toString())
      }
    }),
    EditorView.editable.of(!props.disabled),
    EditorView.theme({
      '&': {
        height: props.height,
        border: '1px solid hsl(var(--border))',
        borderRadius: 'calc(var(--radius) - 2px)',
        backgroundColor: 'hsl(var(--background))',
      },
      '.cm-scroller': { overflow: 'auto' },
      '.cm-content': {
        fontFamily: 'ui-monospace, monospace',
        fontSize: '13px',
        padding: '8px 0',
      },
      '&.cm-focused': {
        outline: '2px solid hsl(var(--ring))',
        outlineOffset: '-1px',
      },
    }),
    ...props.extensions,
  ]
}

onMounted(() => {
  if (!editorContainer.value) return

  const state = EditorState.create({
    doc: props.modelValue,
    extensions: createExtensions(),
  })

  view.value = new EditorView({
    state,
    parent: editorContainer.value,
  })
})

onBeforeUnmount(() => {
  view.value?.destroy()
})

watch(
  () => props.modelValue,
  (newValue) => {
    if (!view.value) return
    const currentValue = view.value.state.doc.toString()
    if (currentValue !== newValue) {
      isUpdating = true
      view.value.dispatch({
        changes: {
          from: 0,
          to: currentValue.length,
          insert: newValue,
        },
      })
      isUpdating = false
    }
  },
)

function insertAtCursor(text: string) {
  if (!view.value) return
  const pos = view.value.state.selection.main.head
  view.value.dispatch({
    changes: { from: pos, insert: text },
    selection: { anchor: pos + text.length },
  })
  view.value.focus()
}

defineExpose({ insertAtCursor })
</script>

<template>
  <div ref="editorContainer" data-testid="codemirror-editor" />
</template>
