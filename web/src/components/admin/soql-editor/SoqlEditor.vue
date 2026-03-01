<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { Extension } from '@codemirror/state'
import { IconButton } from '@/components/ui/icon-button'
import { Code, Type, Check, Play, Braces } from 'lucide-vue-next'
import { Textarea } from '@/components/ui/textarea'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import CodeMirrorEditor from '@/components/admin/expression-builder/CodeMirrorEditor.vue'
import ExpressionErrors from '@/components/admin/expression-builder/ExpressionErrors.vue'
import SoqlObjectPicker from './SoqlObjectPicker.vue'
import SoqlTestResult from './SoqlTestResult.vue'
import { soqlApi, type SoqlValidateError } from '@/api/soql'
import { http } from '@/api/http'
import { soqlLanguage } from '@/lib/codemirror/soql-language'
import { soqlAutocomplete, type SoqlAutocompleteConfig } from '@/lib/codemirror/soql-autocomplete'

interface ObjectInfo {
  apiName: string
  label: string
}

interface FieldInfo {
  apiName: string
  label: string
  fieldType: string
}

interface QueryResult {
  totalSize: number
  records: Record<string, unknown>[]
  error?: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    height?: string
    placeholder?: string
    disabled?: boolean
    showTestQuery?: boolean
  }>(),
  {
    height: '120px',
    placeholder: 'SELECT Id, Name FROM Account WHERE ...',
    disabled: false,
    showTestQuery: true,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const mode = ref<'editor' | 'plain'>('editor')
const validating = ref(false)
const testing = ref(false)
const errors = ref<SoqlValidateError[]>([])
const editorRef = ref<InstanceType<typeof CodeMirrorEditor> | null>(null)
const objects = ref<ObjectInfo[]>([])
const fields = ref<FieldInfo[]>([])
const currentObject = ref<string | null>(null)
const testResult = ref<QueryResult | null>(null)
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

function onEditorBlur() {
  scheduleBlur()
}

function onFocusOut(e: FocusEvent) {
  const container = (e.currentTarget as HTMLElement)
  const related = e.relatedTarget as Node | null
  if (related && container.contains(related)) return
  scheduleBlur()
}

// Extract FROM object from query
const fromObject = computed(() => {
  const match = props.modelValue.match(/\bFROM\s+(\w+)/i)
  return match ? match[1] : null
})

// Load objects list (admin endpoint, no OLS filtering)
async function loadObjects() {
  try {
    const response = await http.get<{ data: ObjectInfo[] }>('/api/v1/admin/soql/objects')
    objects.value = response.data ?? []
  } catch (err) {
    console.error('[SoqlEditor] Failed to load objects:', err)
  }
}

// Load fields for detected FROM object (admin endpoint, no FLS filtering)
async function loadFields(objectName: string) {
  try {
    const response = await http.get<{ data: FieldInfo[] }>(
      `/api/v1/admin/soql/objects/${objectName}/fields`,
    )
    fields.value = response.data ?? []
  } catch {
    fields.value = []
  }
}

// Debounce timer for FROM object detection
let debounceTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => {
  loadObjects()
  if (fromObject.value) {
    currentObject.value = fromObject.value
    loadFields(fromObject.value)
  }
})

// Watch for FROM object changes with debounce to avoid excessive API calls
watch(fromObject, (newObj) => {
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
  const config: SoqlAutocompleteConfig = {
    objects: objects.value,
    fields: fields.value,
  }
  return [soqlAutocomplete(config)]
})

async function onValidate() {
  validating.value = true
  errors.value = []
  try {
    const response = await soqlApi.validate({ query: props.modelValue })
    if (response.valid) {
      errors.value = []
    } else {
      errors.value = response.errors ?? [{ message: 'Query is invalid' }]
    }
  } catch {
    errors.value = [{ message: 'Error validating query' }]
  } finally {
    validating.value = false
  }
}

async function onTestQuery() {
  testing.value = true
  testResult.value = null
  try {
    const response = await http.post<{ totalSize: number; records: Record<string, unknown>[]; error?: string }>(
      '/api/v1/admin/soql/test',
      { query: props.modelValue, pageSize: 5 },
    )
    if (response.error) {
      testResult.value = {
        totalSize: 0,
        records: [],
        error: response.error,
      }
    } else {
      testResult.value = {
        totalSize: response.totalSize ?? 0,
        records: response.records ?? [],
      }
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Query execution failed'
    testResult.value = {
      totalSize: 0,
      records: [],
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
  <div class="space-y-2" data-testid="soql-editor" @focusin="onFocusIn" @focusout="onFocusOut">
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
        data-testid="soql-validate-btn"
        @click="onValidate"
      />
      <IconButton
        v-if="showTestQuery"
        type="button"
        :icon="Play"
        :tooltip="testing ? 'Running...' : 'Test Query'"
        variant="ghost"
        size="icon-sm"
        class="h-7 w-7"
        :disabled="testing || !modelValue || disabled"
        data-testid="soql-test-btn"
        @click="onTestQuery"
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
            data-testid="soql-picker-btn"
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
        FROM {{ currentObject }}
      </span>
    </div>

    <!-- Editor -->
    <CodeMirrorEditor
      v-if="mode === 'editor'"
      ref="editorRef"
      :model-value="modelValue"
      :language="soqlLanguage"
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
      :rows="4"
      :placeholder="placeholder"
      :disabled="disabled"
      class="font-mono text-sm"
      data-testid="soql-textarea"
      @update:model-value="onInput"
    />

    <!-- Test Result -->
    <SoqlTestResult
      v-if="testResult"
      :result="testResult"
      @close="testResult = null"
    />

    <!-- Errors -->
    <ExpressionErrors :errors="errors" @jump-to-error="onJumpToError" />
  </div>
</template>
