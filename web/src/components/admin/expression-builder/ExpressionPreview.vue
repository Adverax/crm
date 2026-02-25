<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { watchDebounced } from '@vueuse/core'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCelEnvironment } from '@/composables/useCelEnvironment'
import type { FunctionParam } from '@/types/functions'
import type { CelEvalResult } from '@/lib/cel'

interface DescribeField {
  apiName: string
  label: string
  fieldType: string
}

const props = withDefaults(
  defineProps<{
    expression: string
    context: string
    functionParams?: FunctionParam[]
    fields?: DescribeField[]
  }>(),
  {
    functionParams: () => [],
    fields: () => [],
  },
)

const emit = defineEmits<{
  'jump-to-position': [position: number]
}>()

const { init, evaluate } = useCelEnvironment()
init()

const result = ref<CelEvalResult | null>(null)
const paramValues = ref<Record<string, string>>({})

const isFunctionBody = computed(() => props.context === 'function_body')

function runEvaluation() {
  if (!props.expression.trim()) {
    result.value = null
    return
  }

  const ctx: Record<string, unknown> = {}

  if (isFunctionBody.value) {
    const hasEmptyParam = props.functionParams.some((p) => !(paramValues.value[p.name] ?? ''))
    if (hasEmptyParam) {
      result.value = null
      return
    }
    for (const p of props.functionParams) {
      ctx[p.name] = coerceParamValue(paramValues.value[p.name] ?? '', p.type)
    }
  } else {
    if (props.fields.length === 0) {
      result.value = null
      return
    }
    const sample = buildSampleRecord(props.fields)
    ctx.record = sample
    ctx.old = sample
    ctx.user = { id: '', profile_id: '', role_id: '' }
    ctx.now = new Date()
  }

  result.value = evaluate(props.expression, ctx)
}

function zeroForType(fieldType: string): unknown {
  switch (fieldType) {
    case 'number':
      return 0
    case 'boolean':
      return false
    case 'datetime':
      return new Date()
    default:
      return ''
  }
}

function buildSampleRecord(fields: DescribeField[]): Record<string, unknown> {
  const base: Record<string, unknown> = {}
  for (const f of fields) {
    base[f.apiName] = zeroForType(f.fieldType)
  }
  // cel-js uses Object.hasOwn() to check key existence in maps,
  // which delegates to getOwnPropertyDescriptor on the Proxy target.
  return new Proxy(base, {
    get(target, prop) {
      if (typeof prop === 'string' && prop in target) {
        return target[prop]
      }
      if (typeof prop === 'string') return 0
      return Reflect.get(target, prop)
    },
    getOwnPropertyDescriptor(target, prop) {
      if (typeof prop === 'string') {
        return (
          Object.getOwnPropertyDescriptor(target, prop) ?? {
            configurable: true,
            enumerable: true,
            value: 0,
          }
        )
      }
      return Reflect.getOwnPropertyDescriptor(target, prop)
    },
    has(_target, prop) {
      if (typeof prop === 'string') return true
      return Reflect.has(_target, prop)
    },
  })
}

function coerceParamValue(raw: string, type: string | undefined): unknown {
  if (!raw) return raw
  switch (type) {
    case 'number': {
      const num = Number(raw)
      if (isNaN(num)) return 0
      // cel-js: integer literals are BigInt, float literals are number.
      // Return BigInt for whole numbers so `2 * x` works (int * int).
      if (Number.isInteger(num)) return BigInt(Math.trunc(num))
      return num
    }
    case 'boolean':
      return raw === 'true'
    default:
      return raw
  }
}

watchDebounced(
  () => [props.expression, paramValues.value, props.fields] as const,
  () => runEvaluation(),
  { debounce: 300, deep: true },
)

watch(
  () => props.functionParams,
  (params) => {
    const newValues: Record<string, string> = {}
    for (const p of params) {
      newValues[p.name] = paramValues.value[p.name] ?? ''
    }
    paramValues.value = newValues
  },
  { immediate: true, deep: true },
)

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return 'null'
  if (typeof value === 'string') return `"${value}"`
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}
</script>

<template>
  <div class="space-y-2" data-testid="expression-preview">
    <!-- Parameter inputs for function_body -->
    <div
      v-if="isFunctionBody && functionParams.length > 0"
      class="flex flex-wrap gap-2"
    >
      <div
        v-for="param in functionParams"
        :key="param.name"
        class="flex items-center gap-1"
      >
        <Label class="text-xs text-muted-foreground whitespace-nowrap">
          {{ param.name }}:
        </Label>
        <Input
          v-model="paramValues[param.name]"
          class="h-7 w-24 text-xs font-mono"
          :placeholder="param.type ?? 'any'"
          :data-testid="`preview-param-${param.name}`"
        />
      </div>
    </div>

    <!-- Result -->
    <div v-if="result" class="text-xs">
      <span
        v-if="result.success"
        class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300 font-mono"
        data-testid="preview-result"
      >
        = {{ formatValue(result.value) }}
        <span class="text-muted-foreground">({{ result.type }})</span>
      </span>
      <span
        v-else
        class="text-destructive font-mono"
        :class="{ 'cursor-pointer hover:underline': result.position != null }"
        data-testid="preview-error"
        :data-clickable="result.position != null ? '' : undefined"
        @click="result.position != null && emit('jump-to-position', result.position)"
      >
        {{ result.error }}
      </span>
    </div>
  </div>
</template>
