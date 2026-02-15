<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Extension } from '@codemirror/state'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import CodeMirrorEditor from './CodeMirrorEditor.vue'
import FieldPicker from './FieldPicker.vue'
import FunctionPicker from './FunctionPicker.vue'
import ExpressionErrors from './ExpressionErrors.vue'
import ExpressionPreview from './ExpressionPreview.vue'
import { celApi } from '@/api/cel'
import { http } from '@/api/http'
import { useFunctionsStore } from '@/stores/functions'
import { celAutocomplete } from '@/lib/codemirror/cel-autocomplete'
import type { CelContext, CelValidateError, FunctionParam } from '@/types/functions'

interface DescribeField {
  apiName: string
  label: string
  fieldType: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    context: CelContext
    objectApiName?: string
    functionParams?: FunctionParam[]
    height?: string
    placeholder?: string
    disabled?: boolean
    showFieldPicker?: boolean
  }>(),
  {
    objectApiName: undefined,
    functionParams: () => [],
    height: '120px',
    placeholder: '',
    disabled: false,
    showFieldPicker: true,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const mode = ref<'editor' | 'plain'>('editor')
const validating = ref(false)
const errors = ref<CelValidateError[]>([])
const returnType = ref<string | null>(null)
const fields = ref<DescribeField[]>([])
const editorRef = ref<InstanceType<typeof CodeMirrorEditor> | null>(null)
const showPreview = ref(false)
const pickerTab = ref('fields')

const functionsStore = useFunctionsStore()

async function loadFields() {
  if (!props.objectApiName) return
  try {
    const response = await http.get<{ data: { fields: DescribeField[] } }>(
      `/api/v1/describe/${props.objectApiName}`,
    )
    fields.value = (response.data.fields ?? []).filter(
      (f: DescribeField) => !['Id', 'CreatedAt', 'UpdatedAt', 'OwnerId', 'CreatedById', 'UpdatedById'].includes(f.apiName),
    )
  } catch {
    // Silently fail — fields won't be shown in picker
  }
}

onMounted(() => {
  functionsStore.ensureLoaded()
  if (props.objectApiName) {
    loadFields()
  }
})

const autocompleteExtension = computed<Extension[]>(() => {
  return [
    celAutocomplete({
      fields: fields.value,
      params: props.functionParams,
      functions: functionsStore.functions,
      context: props.context,
    }),
  ]
})

async function onValidate() {
  validating.value = true
  errors.value = []
  returnType.value = null
  try {
    const response = await celApi.validate({
      expression: props.modelValue,
      context: props.context,
      objectApiName: props.objectApiName,
      params: props.functionParams?.map((p) => ({ name: p.name, type: p.type })),
    })
    if (response.valid) {
      returnType.value = response.returnType ?? null
      errors.value = []
    } else {
      errors.value = response.errors ?? [{ message: 'Выражение невалидно' }]
    }
  } catch {
    errors.value = [{ message: 'Ошибка при валидации выражения' }]
  } finally {
    validating.value = false
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

const showSidePanel = computed(() => {
  if (!props.showFieldPicker) return false
  return fields.value.length > 0 ||
    props.functionParams.length > 0 ||
    functionsStore.functions.length > 0 ||
    props.context !== 'function_body'
})

const hasFunctions = computed(() => functionsStore.functions.length > 0)
</script>

<template>
  <div class="space-y-2" data-testid="expression-builder">
    <!-- Toolbar -->
    <div class="flex items-center gap-2">
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="h-7 px-2 text-xs"
        @click="mode = mode === 'editor' ? 'plain' : 'editor'"
      >
        {{ mode === 'editor' ? 'Текст' : 'Редактор' }}
      </Button>
      <Button
        type="button"
        variant="outline"
        size="sm"
        class="h-7 px-2 text-xs"
        :disabled="validating || !modelValue"
        data-testid="validate-btn"
        @click="onValidate"
      >
        {{ validating ? 'Проверка...' : 'Проверить' }}
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="h-7 px-2 text-xs"
        data-testid="preview-toggle"
        @click="showPreview = !showPreview"
      >
        {{ showPreview ? 'Скрыть превью' : 'Превью' }}
      </Button>
      <span
        v-if="returnType"
        class="text-xs text-muted-foreground"
        data-testid="return-type"
      >
        → {{ returnType }}
      </span>
    </div>

    <!-- Editor + Side Panel -->
    <div class="flex gap-3">
      <div class="flex-1 min-w-0">
        <CodeMirrorEditor
          v-if="mode === 'editor'"
          ref="editorRef"
          :model-value="modelValue"
          :extensions="autocompleteExtension"
          :height="height"
          :disabled="disabled"
          @update:model-value="onInput"
        />
        <Textarea
          v-else
          :model-value="modelValue"
          :rows="4"
          :placeholder="placeholder"
          :disabled="disabled"
          class="font-mono text-sm"
          data-testid="expression-textarea"
          @update:model-value="onInput"
        />
      </div>

      <!-- Side panel with tabs -->
      <div v-if="showSidePanel" class="w-60 border-l pl-3">
        <Tabs v-model="pickerTab" class="w-full">
          <TabsList class="h-8 w-full" data-testid="picker-tabs">
            <TabsTrigger value="fields" class="text-xs h-7 flex-1">
              Поля
            </TabsTrigger>
            <TabsTrigger
              v-if="hasFunctions || context !== 'function_body'"
              value="functions"
              class="text-xs h-7 flex-1"
              data-testid="functions-tab"
            >
              Функции
            </TabsTrigger>
          </TabsList>

          <TabsContent value="fields" class="mt-2">
            <FieldPicker
              :fields="fields"
              :params="functionParams"
              :context="context"
              class="!w-full !border-0 !pl-0"
              @insert="onInsertFromPicker"
            />
          </TabsContent>

          <TabsContent value="functions" class="mt-2" data-testid="functions-tab-content">
            <FunctionPicker @insert="onInsertFromPicker" />
          </TabsContent>
        </Tabs>
      </div>
    </div>

    <!-- Preview -->
    <ExpressionPreview
      v-if="showPreview"
      :expression="modelValue"
      :context="context"
      :function-params="functionParams"
    />

    <!-- Errors -->
    <ExpressionErrors :errors="errors" />
  </div>
</template>
