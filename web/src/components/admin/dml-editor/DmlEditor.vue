<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { Extension } from '@codemirror/state'
import { IconButton } from '@/components/ui/icon-button'
import { Code, Type, Check, Play, Braces } from 'lucide-vue-next'
import { Textarea } from '@/components/ui/textarea'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import CodeMirrorEditor from '@/components/admin/expression-builder/CodeMirrorEditor.vue'
import ExpressionErrors from '@/components/admin/expression-builder/ExpressionErrors.vue'
import SoqlObjectPicker from '@/components/admin/soql-editor/SoqlObjectPicker.vue'
import DmlTestResult from './DmlTestResult.vue'
import { dmlApi, type DmlValidateError, type DmlTestResponse } from '@/api/dml'
import { http } from '@/api/http'
import { dmlLanguage } from '@/lib/codemirror/dml-language'
import { dmlAutocomplete, type DmlAutocompleteConfig } from '@/lib/codemirror/dml-autocomplete'

interface ObjectInfo {
  apiName: string
  label: string
}

interface FieldInfo {
  apiName: string
  label: string
  fieldType: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    height?: string
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    height: '80px',
    placeholder: "INSERT INTO Account (Name) VALUES ('Acme')",
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const mode = ref<'editor' | 'plain'>('editor')
const validating = ref(false)
const testing = ref(false)
const errors = ref<DmlValidateError[]>([])
const editorRef = ref<InstanceType<typeof CodeMirrorEditor> | null>(null)
const objects = ref<ObjectInfo[]>([])
const fields = ref<FieldInfo[]>([])
const currentObject = ref<string | null>(null)
const testResult = ref<DmlTestResponse | null>(null)
const focused = ref(false)
const pickerOpen = ref(false)
const toolbarVisible = computed(() => focused.value || pickerOpen.value)

function onFocusIn() {
  cancelBlur()
  focused.value = true
}

let blurTimer: ReturnType<typeof setTimeout> | null = null

function cancelBlur() {
  if (blurTimer) {
    clearTimeout(blurTimer)
    blurTimer = null
  }
}

function scheduleBlur() {
  cancelBlur()
  blurTimer = setTimeout(() => {
    focused.value = false
    testResult.value = null
  }, 100)
}

// When mode toggles, the old editor unmounts → blur/focusout fires → scheduleBlur().
// Cancel that timer after DOM update to keep toolbar visible.
watch(mode, () => {
  cancelBlur()
  focused.value = true
}, { flush: 'post' })

function onEditorBlur() {
  scheduleBlur()
}

function onFocusOut(e: FocusEvent) {
  const container = (e.currentTarget as HTMLElement)
  const related = e.relatedTarget as Node | null
  if (related && container.contains(related)) return
  scheduleBlur()
}

// Extract target object from DML statement
const dmlObject = computed(() => {
  const val = props.modelValue
  const insertMatch = val.match(/\bINSERT\s+INTO\s+(\w+)/i)
  if (insertMatch) return insertMatch[1]!
  const updateMatch = val.match(/\bUPDATE\s+(\w+)/i)
  if (updateMatch) return updateMatch[1]!
  const deleteMatch = val.match(/\bDELETE\s+FROM\s+(\w+)/i)
  if (deleteMatch) return deleteMatch[1]!
  const upsertMatch = val.match(/\bUPSERT\s+(\w+)/i)
  if (upsertMatch) return upsertMatch[1]!
  return null
})

// Load objects list (admin endpoint, no OLS filtering)
async function loadObjects() {
  try {
    const response = await http.get<{ data: ObjectInfo[] }>('/api/v1/admin/dml/objects')
    objects.value = response.data ?? []
  } catch (err) {
    console.error('[DmlEditor] Failed to load objects:', err)
  }
}

// Load fields for detected object (admin endpoint, no FLS filtering)
async function loadFields(objectName: string) {
  try {
    const response = await http.get<{ data: FieldInfo[] }>(
      `/api/v1/admin/dml/objects/${objectName}/fields`,
    )
    fields.value = response.data ?? []
  } catch {
    fields.value = []
  }
}

// Debounce timer for object detection
let debounceTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => {
  loadObjects()
  if (dmlObject.value) {
    currentObject.value = dmlObject.value
    loadFields(dmlObject.value)
  }
})

// Watch for object changes with debounce
watch(dmlObject, (newObj) => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    if (newObj && newObj !== currentObject.value) {
      currentObject.value = newObj
      loadFields(newObj)
    }
  }, 500)
})

// Autocomplete config
const autocompleteExtension = computed<Extension[]>(() => {
  const config: DmlAutocompleteConfig = {
    objects: objects.value,
    fields: fields.value,
  }
  return [dmlAutocomplete(config)]
})

async function onValidate() {
  validating.value = true
  errors.value = []
  try {
    const response = await dmlApi.validate({ statement: props.modelValue })
    if (response.valid) {
      errors.value = []
    } else {
      errors.value = response.errors ?? [{ message: 'Statement is invalid' }]
    }
  } catch {
    errors.value = [{ message: 'Error validating statement' }]
  } finally {
    validating.value = false
  }
}

async function onTestExecute() {
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await dmlApi.test({ statement: props.modelValue })
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Statement execution failed'
    testResult.value = {
      operation: '',
      object: '',
      rowsAffected: 0,
      rolledBack: true,
      error: message,
    }
  } finally {
    testing.value = false
  }
}

function onSelectObject(objectName: string) {
  if (objectName !== currentObject.value) {
    currentObject.value = objectName
    loadFields(objectName)
  }
}

function onInsertFromPicker(text: string) {
  if (mode.value === 'editor' && editorRef.value) {
    editorRef.value.insertAtCursor(text)
  } else {
    emit('update:modelValue', props.modelValue + text)
  }
}

function onInput(value: string | number) {
  emit('update:modelValue', String(value))
}

function onJumpToError(line: number, column: number) {
  if (mode.value === 'editor' && editorRef.value) {
    editorRef.value.setCursorAtLineCol(line, column)
  }
}
</script>

<template>
  <div class="space-y-2" data-testid="dml-editor" @focusin="onFocusIn" @focusout="onFocusOut">
    <!-- Toolbar -->
    <div v-show="toolbarVisible" class="flex items-center gap-1" @mousedown.prevent>
      <IconButton
        type="button"
        :icon="mode === 'editor' ? Type : Code"
        :tooltip="mode === 'editor' ? 'Switch to plain text' : 'Switch to editor'"
        variant="ghost"
        size="icon-sm"
        class="h-7 w-7"
        @click="mode = mode === 'editor' ? 'plain' : 'editor'"
      />
      <IconButton
        type="button"
        :icon="Check"
        :tooltip="validating ? 'Validating...' : 'Validate'"
        variant="outline"
        size="icon-sm"
        class="h-7 w-7"
        :disabled="validating || !modelValue || disabled"
        data-testid="dml-validate-btn"
        @click="onValidate"
      />
      <IconButton
        type="button"
        :icon="Play"
        :tooltip="testing ? 'Running...' : 'Test Execute'"
        variant="ghost"
        size="icon-sm"
        class="h-7 w-7"
        :disabled="testing || !modelValue || disabled"
        data-testid="dml-test-btn"
        @click="onTestExecute"
      />
      <Popover @update:open="(open: boolean) => { pickerOpen = open; if (open && objects.length === 0) loadObjects() }">
        <PopoverTrigger as-child>
          <IconButton
            type="button"
            :icon="Braces"
            tooltip="Objects & Fields"
            variant="ghost"
            size="icon-sm"
            class="h-7 w-7"
            data-testid="dml-picker-btn"
          />
        </PopoverTrigger>
        <PopoverContent
          side="bottom"
          align="end"
          class="w-72 p-3"
          @open-auto-focus.prevent
          @focus-outside.prevent
        >
          <SoqlObjectPicker
            :objects="objects"
            :fields="fields"
            :current-object="currentObject"
            @insert="onInsertFromPicker"
            @select-object="onSelectObject"
          />
        </PopoverContent>
      </Popover>
      <span
        v-if="currentObject"
        class="text-xs text-muted-foreground"
      >
        {{ currentObject }}
      </span>
    </div>

    <!-- Editor -->
    <CodeMirrorEditor
      v-if="mode === 'editor'"
      ref="editorRef"
      :model-value="modelValue"
      :language="dmlLanguage"
      :extensions="autocompleteExtension"
      :height="height"
      :disabled="disabled"
      @update:model-value="onInput"
      @focus="onFocusIn"
      @blur="onEditorBlur"
    />
    <Textarea
      v-else
      :model-value="modelValue"
      :rows="2"
      :placeholder="placeholder"
      :disabled="disabled"
      class="font-mono text-sm"
      data-testid="dml-textarea"
      @update:model-value="onInput"
    />

    <!-- Test Result -->
    <DmlTestResult
      v-if="testResult"
      :result="testResult"
      @close="testResult = null"
    />

    <!-- Errors -->
    <ExpressionErrors :errors="errors" @jump-to-error="onJumpToError" />
  </div>
</template>
