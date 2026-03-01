<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState, Compartment, type Extension } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { bracketMatching } from '@codemirror/language'

const props = withDefaults(
  defineProps<{
    modelValue: string
    language?: Extension
    extensions?: Extension[]
    height?: string
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    language: undefined,
    extensions: () => [],
    height: '120px',
    placeholder: '',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focus: []
  blur: []
}>()

const editorContainer = ref<HTMLDivElement | null>(null)
const view = shallowRef<EditorView | null>(null)
const dynamicCompartment = new Compartment()
let isUpdating = false

function createExtensions(): Extension[] {
  const result: Extension[] = []
  if (props.language) {
    result.push(props.language)
  }
  return [
    ...result,
    history(),
    bracketMatching(),
    keymap.of([...defaultKeymap, ...historyKeymap]),
    EditorView.lineWrapping,
    EditorView.updateListener.of((update) => {
      if (update.docChanged && !isUpdating) {
        emit('update:modelValue', update.state.doc.toString())
      }
      if (update.focusChanged) {
        if (update.view.hasFocus) {
          emit('focus')
        } else {
          emit('blur')
        }
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
    dynamicCompartment.of(props.extensions),
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

watch(
  () => props.extensions,
  (newExtensions) => {
    if (!view.value) return
    view.value.dispatch({
      effects: dynamicCompartment.reconfigure(newExtensions),
    })
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

function setCursorAt(offset: number) {
  if (!view.value) return
  const docLength = view.value.state.doc.length
  const pos = Math.min(Math.max(0, offset), docLength)
  view.value.dispatch({ selection: { anchor: pos } })
  view.value.focus()
}

function setCursorAtLineCol(line: number, col: number) {
  if (!view.value) return
  const doc = view.value.state.doc
  const lineNum = Math.min(Math.max(1, line), doc.lines)
  const lineObj = doc.line(lineNum)
  const offset = Math.min(lineObj.from + Math.max(0, col - 1), lineObj.to)
  view.value.dispatch({ selection: { anchor: offset } })
  view.value.focus()
}

defineExpose({ insertAtCursor, setCursorAt, setCursorAtLineCol })
</script>

<template>
  <div ref="editorContainer" data-testid="codemirror-editor" />
</template>
