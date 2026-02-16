import {
  autocompletion,
  type CompletionContext,
  type CompletionResult,
  type Completion,
} from '@codemirror/autocomplete'
import type { Extension } from '@codemirror/state'
import type { Function, FunctionParam } from '@/types/functions'

interface FieldInfo {
  apiName: string
  label: string
  fieldType: string
}

interface AutocompleteConfig {
  fields?: FieldInfo[]
  params?: FunctionParam[]
  functions?: Function[]
  context: string
}

const builtinFunctions: Completion[] = [
  { label: 'size', detail: '(value): int', type: 'function', info: 'Length of string, list, or map' },
  { label: 'contains', detail: '(substr): bool', type: 'function', info: 'Checks if string contains substring' },
  { label: 'startsWith', detail: '(prefix): bool', type: 'function', info: 'Checks if string starts with prefix' },
  { label: 'endsWith', detail: '(suffix): bool', type: 'function', info: 'Checks if string ends with suffix' },
  { label: 'matches', detail: '(regex): bool', type: 'function', info: 'Matches against regular expression' },
  { label: 'int', detail: '(value): int', type: 'function', info: 'Convert to integer' },
  { label: 'double', detail: '(value): double', type: 'function', info: 'Convert to floating point number' },
  { label: 'string', detail: '(value): string', type: 'function', info: 'Convert to string' },
  { label: 'bool', detail: '(value): bool', type: 'function', info: 'Convert to boolean' },
  { label: 'duration', detail: '(string): duration', type: 'function', info: 'Create a duration' },
  { label: 'timestamp', detail: '(string): timestamp', type: 'function', info: 'Create a timestamp' },
  { label: 'has', detail: '(field): bool', type: 'keyword', info: 'Checks if field exists' },
  { label: 'type', detail: '(value): string', type: 'function', info: 'Returns the type of value' },
]

const keywords: Completion[] = [
  { label: 'true', type: 'keyword' },
  { label: 'false', type: 'keyword' },
  { label: 'null', type: 'keyword' },
  { label: 'in', type: 'keyword', detail: 'membership operator' },
]

const operators: Completion[] = [
  { label: '&&', detail: 'logical AND', type: 'keyword' },
  { label: '||', detail: 'logical OR', type: 'keyword' },
  { label: '==', detail: 'equals', type: 'keyword' },
  { label: '!=', detail: 'not equals', type: 'keyword' },
  { label: '>=', detail: 'greater than or equal', type: 'keyword' },
  { label: '<=', detail: 'less than or equal', type: 'keyword' },
  { label: '?:', detail: 'ternary operator', type: 'keyword' },
]

function buildFieldCompletions(
  prefix: string,
  fields: FieldInfo[],
): Completion[] {
  return fields.map((f) => ({
    label: `${prefix}.${f.apiName}`,
    detail: f.fieldType,
    info: f.label,
    type: 'property',
  }))
}

function buildFunctionCompletions(functions: Function[]): Completion[] {
  return functions
    .filter((fn) => !!fn.name)
    .map((fn) => {
      const params = fn.params ?? []
      const paramStr = params.map((p) => p.name).join(', ')
      const sig = `(${paramStr}): ${fn.returnType ?? 'any'}`
      return {
        label: `fn.${fn.name!}`,
        detail: sig,
        info: fn.description ?? undefined,
        type: 'function',
        apply: `fn.${fn.name!}(${paramStr})`,
      }
    })
}

function buildParamCompletions(params: FunctionParam[]): Completion[] {
  return params.map((p) => ({
    label: p.name,
    detail: p.type ?? 'any',
    info: p.description ?? undefined,
    type: 'variable',
  }))
}

function celCompletionSource(config: AutocompleteConfig) {
  return (ctx: CompletionContext): CompletionResult | null => {
    const { fields = [], params = [], functions = [] } = config
    const isFunctionBody = config.context === 'function_body'

    // Check what's before the cursor
    const line = ctx.state.doc.lineAt(ctx.pos)
    const textBefore = line.text.slice(0, ctx.pos - line.from)

    // After "record." — show record fields
    const recordMatch = textBefore.match(/record\.(\w*)$/)
    if (recordMatch && !isFunctionBody && fields.length > 0) {
      const from = ctx.pos - (recordMatch[1]?.length ?? 0)
      const completions = fields.map((f) => ({
        label: f.apiName,
        detail: f.fieldType,
        info: f.label,
        type: 'property' as const,
      }))
      return { from, options: completions }
    }

    // After "old." — show fields (only for validation_rule/when_expression)
    const oldMatch = textBefore.match(/old\.(\w*)$/)
    if (
      oldMatch &&
      (config.context === 'validation_rule' || config.context === 'when_expression')
    ) {
      const from = ctx.pos - (oldMatch[1]?.length ?? 0)
      const completions = fields.map((f) => ({
        label: f.apiName,
        detail: f.fieldType,
        info: f.label,
        type: 'property' as const,
      }))
      return { from, options: completions }
    }

    // After "user." — show user fields
    const userMatch = textBefore.match(/user\.(\w*)$/)
    if (userMatch && !isFunctionBody) {
      const from = ctx.pos - (userMatch[1]?.length ?? 0)
      return {
        from,
        options: [
          { label: 'id', detail: 'string', info: 'User ID', type: 'property' },
          { label: 'profile_id', detail: 'string', info: 'Profile ID', type: 'property' },
          { label: 'role_id', detail: 'string', info: 'Role ID', type: 'property' },
        ],
      }
    }

    // After "fn." — show custom functions
    const fnMatch = textBefore.match(/fn\.(\w*)$/)
    if (fnMatch) {
      const from = ctx.pos - (fnMatch[1]?.length ?? 0)
      const completions: Completion[] = functions
        .filter((fn): fn is typeof fn & { name: string } => !!fn.name)
        .map((fn) => {
          const fnParams = fn.params ?? []
          const paramStr = fnParams.map((p) => p.name).join(', ')
          return {
            label: fn.name,
            detail: `(${paramStr}): ${fn.returnType ?? 'any'}`,
            info: fn.description ?? undefined,
            type: 'function' as const,
            apply: `${fn.name}(${paramStr})`,
          }
        })
      return { from, options: completions }
    }

    // General word completion
    const wordMatch = textBefore.match(/(\w+)$/)
    if (!wordMatch && !ctx.explicit) return null

    const from = wordMatch ? ctx.pos - (wordMatch[1]?.length ?? 0) : ctx.pos

    const options: Completion[] = []

    // Context variables
    if (!isFunctionBody) {
      if (fields.length > 0) {
        options.push({ label: 'record', type: 'variable', detail: 'record fields', boost: 2 })
      }
      if (config.context === 'validation_rule' || config.context === 'when_expression') {
        options.push({ label: 'old', type: 'variable', detail: 'previous values', boost: 2 })
      }
      options.push({ label: 'user', type: 'variable', detail: 'current user', boost: 1 })
      options.push({ label: 'now', type: 'variable', detail: 'current time' })
    }

    // Function params
    if (isFunctionBody && params.length > 0) {
      options.push(...buildParamCompletions(params))
    }

    // Record/old field completions (full path for general context)
    if (!isFunctionBody && fields.length > 0) {
      options.push(...buildFieldCompletions('record', fields))
      if (config.context === 'validation_rule' || config.context === 'when_expression') {
        options.push(...buildFieldCompletions('old', fields))
      }
    }

    // fn.* namespace
    if (functions.length > 0) {
      options.push({ label: 'fn', type: 'namespace', detail: 'custom functions', boost: 1 })
      options.push(...buildFunctionCompletions(functions))
    }

    // Built-in functions
    options.push(...builtinFunctions)

    // Keywords and operators
    options.push(...keywords)
    options.push(...operators)

    return { from, options, validFor: /^\w*$/ }
  }
}

export function celAutocomplete(config: AutocompleteConfig): Extension {
  return autocompletion({
    override: [celCompletionSource(config)],
    activateOnTyping: true,
  })
}
