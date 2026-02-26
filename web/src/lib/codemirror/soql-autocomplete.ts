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

export interface SoqlAutocompleteConfig {
  objects?: ObjectInfo[]
  fields?: FieldInfo[]
}

const soqlKeywords: Completion[] = [
  { label: 'SELECT', type: 'keyword' },
  { label: 'FROM', type: 'keyword' },
  { label: 'WHERE', type: 'keyword' },
  { label: 'AND', type: 'keyword' },
  { label: 'OR', type: 'keyword' },
  { label: 'NOT', type: 'keyword' },
  { label: 'IN', type: 'keyword' },
  { label: 'LIKE', type: 'keyword' },
  { label: 'IS', type: 'keyword' },
  { label: 'NULL', type: 'keyword' },
  { label: 'ORDER BY', type: 'keyword' },
  { label: 'GROUP BY', type: 'keyword' },
  { label: 'HAVING', type: 'keyword' },
  { label: 'LIMIT', type: 'keyword' },
  { label: 'OFFSET', type: 'keyword' },
  { label: 'ASC', type: 'keyword' },
  { label: 'DESC', type: 'keyword' },
  { label: 'NULLS FIRST', type: 'keyword' },
  { label: 'NULLS LAST', type: 'keyword' },
  { label: 'TRUE', type: 'keyword' },
  { label: 'FALSE', type: 'keyword' },
  { label: 'WITH SECURITY_ENFORCED', type: 'keyword' },
]

const aggregateFunctions: Completion[] = [
  { label: 'COUNT()', type: 'function', detail: 'Count records' },
  { label: 'COUNT_DISTINCT(', type: 'function', detail: 'Count distinct values' },
  { label: 'SUM(', type: 'function', detail: 'Sum of values' },
  { label: 'AVG(', type: 'function', detail: 'Average of values' },
  { label: 'MIN(', type: 'function', detail: 'Minimum value' },
  { label: 'MAX(', type: 'function', detail: 'Maximum value' },
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

const dateLiterals: Completion[] = [
  { label: 'TODAY', type: 'constant', detail: 'Current date' },
  { label: 'YESTERDAY', type: 'constant', detail: 'Previous date' },
  { label: 'TOMORROW', type: 'constant', detail: 'Next date' },
  { label: 'THIS_WEEK', type: 'constant', detail: 'Current week' },
  { label: 'LAST_WEEK', type: 'constant', detail: 'Previous week' },
  { label: 'THIS_MONTH', type: 'constant', detail: 'Current month' },
  { label: 'LAST_MONTH', type: 'constant', detail: 'Previous month' },
  { label: 'THIS_QUARTER', type: 'constant', detail: 'Current quarter' },
  { label: 'THIS_YEAR', type: 'constant', detail: 'Current year' },
  { label: 'LAST_YEAR', type: 'constant', detail: 'Previous year' },
  { label: 'LAST_N_DAYS:', type: 'constant', detail: 'Last N days' },
  { label: 'NEXT_N_DAYS:', type: 'constant', detail: 'Next N days' },
]

type SoqlContext = 'select' | 'from' | 'where' | 'orderby' | 'groupby' | 'general'

function detectContext(textBefore: string): SoqlContext {
  const upper = textBefore.toUpperCase()

  // Find the last significant keyword
  const lastFrom = upper.lastIndexOf('FROM')
  const lastSelect = upper.lastIndexOf('SELECT')
  const lastWhere = upper.lastIndexOf('WHERE')
  const lastOrderBy = upper.lastIndexOf('ORDER BY')
  const lastGroupBy = upper.lastIndexOf('GROUP BY')

  const positions = [
    { ctx: 'select' as SoqlContext, pos: lastSelect },
    { ctx: 'from' as SoqlContext, pos: lastFrom },
    { ctx: 'where' as SoqlContext, pos: lastWhere },
    { ctx: 'orderby' as SoqlContext, pos: lastOrderBy },
    { ctx: 'groupby' as SoqlContext, pos: lastGroupBy },
  ].filter((p) => p.pos >= 0)

  if (positions.length === 0) return 'general'

  positions.sort((a, b) => b.pos - a.pos)
  return positions[0]!.ctx
}

function soqlCompletionSource(config: SoqlAutocompleteConfig) {
  return (ctx: CompletionContext): CompletionResult | null => {
    const { objects = [], fields = [] } = config

    // Get text before cursor (full document up to cursor)
    const doc = ctx.state.doc
    const textBefore = doc.sliceString(0, ctx.pos)

    // Get the current word
    const line = doc.lineAt(ctx.pos)
    const lineText = line.text.slice(0, ctx.pos - line.from)
    const wordMatch = lineText.match(/(\w+)$/)
    if (!wordMatch && !ctx.explicit) return null

    const from = wordMatch ? ctx.pos - (wordMatch[1]?.length ?? 0) : ctx.pos
    const soqlCtx = detectContext(textBefore)

    const options: Completion[] = []

    switch (soqlCtx) {
      case 'select': {
        // After SELECT or , in SELECT clause: fields, *, aggregate functions
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(...aggregateFunctions)
        options.push(...scalarFunctions)
        // Add FROM to move to next clause
        options.push({ label: 'FROM', type: 'keyword' })
        break
      }
      case 'from': {
        // After FROM: object names
        const objectCompletions = objects.map((o) => ({
          label: o.apiName,
          detail: o.label,
          type: 'class' as const,
        }))
        options.push(...objectCompletions)
        break
      }
      case 'where': {
        // After WHERE/AND/OR: fields, parameters, date literals, operators
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(...dateLiterals)
        options.push(
          { label: 'AND', type: 'keyword' },
          { label: 'OR', type: 'keyword' },
          { label: 'NOT', type: 'keyword' },
          { label: 'IN', type: 'keyword' },
          { label: 'LIKE', type: 'keyword' },
          { label: 'IS NULL', type: 'keyword' },
          { label: 'IS NOT NULL', type: 'keyword' },
          { label: 'ORDER BY', type: 'keyword' },
          { label: 'GROUP BY', type: 'keyword' },
          { label: 'LIMIT', type: 'keyword' },
        )
        options.push(...scalarFunctions)
        break
      }
      case 'orderby': {
        // After ORDER BY: fields, ASC, DESC
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(
          { label: 'ASC', type: 'keyword' },
          { label: 'DESC', type: 'keyword' },
          { label: 'NULLS FIRST', type: 'keyword' },
          { label: 'NULLS LAST', type: 'keyword' },
          { label: 'LIMIT', type: 'keyword' },
        )
        break
      }
      case 'groupby': {
        // After GROUP BY: fields
        const fieldCompletions = fields.map((f) => ({
          label: f.apiName,
          detail: `${f.fieldType} — ${f.label}`,
          type: 'property' as const,
        }))
        options.push(...fieldCompletions)
        options.push(
          { label: 'HAVING', type: 'keyword' },
          { label: 'ORDER BY', type: 'keyword' },
          { label: 'LIMIT', type: 'keyword' },
        )
        break
      }
      default: {
        // General context: all keywords
        options.push(...soqlKeywords)
        break
      }
    }

    return { from, options, validFor: /^\w*$/ }
  }
}

export function soqlAutocomplete(config: SoqlAutocompleteConfig): Extension {
  return autocompletion({
    override: [soqlCompletionSource(config)],
    activateOnTyping: true,
  })
}
