import {
  autocompletion,
  type CompletionContext,
  type CompletionResult,
  type Completion,
} from '@codemirror/autocomplete'
import type { Extension } from '@codemirror/state'

interface ObjectInfo {
  apiName: string
  label: string
}

interface FieldInfo {
  apiName: string
  label: string
  fieldType: string
}

export interface DmlAutocompleteConfig {
  objects?: ObjectInfo[]
  fields?: FieldInfo[]
}

const dmlKeywords: Completion[] = [
  { label: 'INSERT INTO', type: 'keyword' },
  { label: 'VALUES', type: 'keyword' },
  { label: 'UPDATE', type: 'keyword' },
  { label: 'SET', type: 'keyword' },
  { label: 'DELETE FROM', type: 'keyword' },
  { label: 'UPSERT', type: 'keyword' },
  { label: 'ON', type: 'keyword' },
  { label: 'WHERE', type: 'keyword' },
  { label: 'AND', type: 'keyword' },
  { label: 'OR', type: 'keyword' },
  { label: 'NOT', type: 'keyword' },
  { label: 'IN', type: 'keyword' },
  { label: 'LIKE', type: 'keyword' },
  { label: 'IS NULL', type: 'keyword' },
  { label: 'IS NOT NULL', type: 'keyword' },
  { label: 'TRUE', type: 'keyword' },
  { label: 'FALSE', type: 'keyword' },
]

const scalarFunctions: Completion[] = [
  { label: 'COALESCE(', type: 'function', detail: 'First non-null value' },
  { label: 'NULLIF(', type: 'function', detail: 'Returns null if equal' },
  { label: 'CONCAT(', type: 'function', detail: 'Concatenate strings' },
  { label: 'UPPER(', type: 'function', detail: 'Uppercase string' },
  { label: 'LOWER(', type: 'function', detail: 'Lowercase string' },
  { label: 'TRIM(', type: 'function', detail: 'Trim whitespace' },
  { label: 'LENGTH(', type: 'function', detail: 'String length' },
  { label: 'ABS(', type: 'function', detail: 'Absolute value' },
  { label: 'ROUND(', type: 'function', detail: 'Round number' },
]

type DmlContext = 'insert_into' | 'insert_fields' | 'values' | 'update_object' | 'set' | 'delete_from' | 'where' | 'general'

function detectContext(textBefore: string): DmlContext {
  const upper = textBefore.toUpperCase()

  const lastInsertInto = upper.lastIndexOf('INSERT INTO')
  const lastValues = upper.lastIndexOf('VALUES')
  const lastUpdate = upper.lastIndexOf('UPDATE')
  const lastSet = upper.lastIndexOf('SET')
  const lastDeleteFrom = upper.lastIndexOf('DELETE FROM')
  const lastWhere = upper.lastIndexOf('WHERE')

  const positions = [
    { ctx: 'insert_into' as DmlContext, pos: lastInsertInto },
    { ctx: 'values' as DmlContext, pos: lastValues },
    { ctx: 'update_object' as DmlContext, pos: lastUpdate },
    { ctx: 'set' as DmlContext, pos: lastSet },
    { ctx: 'delete_from' as DmlContext, pos: lastDeleteFrom },
    { ctx: 'where' as DmlContext, pos: lastWhere },
  ].filter((p) => p.pos >= 0)

  if (positions.length === 0) return 'general'

  positions.sort((a, b) => b.pos - a.pos)
  const latest = positions[0]!

  // After INSERT INTO: check if we're past the object name (in field list)
  if (latest.ctx === 'insert_into') {
    const afterInsert = upper.slice(latest.pos + 'INSERT INTO'.length).trim()
    // If there's already a word (object name) and a `(`, we're in field list
    if (/^\w+\s*\(/.test(afterInsert)) {
      return 'insert_fields'
    }
    return 'insert_into'
  }

  return latest.ctx
}

function dmlCompletionSource(config: DmlAutocompleteConfig) {
  return (ctx: CompletionContext): CompletionResult | null => {
    const { objects = [], fields = [] } = config

    const doc = ctx.state.doc
    const textBefore = doc.sliceString(0, ctx.pos)

    const line = doc.lineAt(ctx.pos)
    const lineText = line.text.slice(0, ctx.pos - line.from)
    const wordMatch = lineText.match(/(\w+)$/)
    if (!wordMatch && !ctx.explicit) return null

    const from = wordMatch ? ctx.pos - (wordMatch[1]?.length ?? 0) : ctx.pos
    const dmlCtx = detectContext(textBefore)

    const options: Completion[] = []

    switch (dmlCtx) {
      case 'insert_into':
      case 'update_object':
      case 'delete_from': {
        // After INSERT INTO / UPDATE / DELETE FROM: object names
        const objectCompletions = objects.map((o) => ({
          label: o.apiName,
          detail: o.label,
          type: 'class' as const,
        }))
        options.push(...objectCompletions)
        break
      }
      case 'insert_fields': {
        // After INSERT INTO Object (: field names
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        break
      }
      case 'values': {
        // After VALUES (: functions and keywords
        options.push(...scalarFunctions)
        options.push(
          { label: 'NULL', type: 'keyword' },
          { label: 'TRUE', type: 'keyword' },
          { label: 'FALSE', type: 'keyword' },
        )
        break
      }
      case 'set': {
        // After SET: field names, functions
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(...scalarFunctions)
        options.push({ label: 'WHERE', type: 'keyword' })
        break
      }
      case 'where': {
        // After WHERE: field names, operators, functions
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(
          { label: 'AND', type: 'keyword' },
          { label: 'OR', type: 'keyword' },
          { label: 'NOT', type: 'keyword' },
          { label: 'IN', type: 'keyword' },
          { label: 'LIKE', type: 'keyword' },
          { label: 'IS NULL', type: 'keyword' },
          { label: 'IS NOT NULL', type: 'keyword' },
        )
        options.push(...scalarFunctions)
        break
      }
      default: {
        // General context: all keywords
        options.push(...dmlKeywords)
        break
      }
    }

    return { from, options, validFor: /^\w*$/ }
  }
}

export function dmlAutocomplete(config: DmlAutocompleteConfig): Extension {
  return autocompletion({
    override: [dmlCompletionSource(config)],
    activateOnTyping: true,
  })
}
