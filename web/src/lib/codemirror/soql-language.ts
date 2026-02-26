import { StreamLanguage, type StreamParser } from '@codemirror/language'
import { tags as t } from '@lezer/highlight'

interface SoqlState {
  inString: boolean
  inComment: boolean
}

const keywords = new Set([
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'LIKE',
  'IS', 'GROUP', 'BY', 'HAVING', 'ORDER', 'LIMIT', 'OFFSET',
  'ASC', 'DESC', 'NULLS', 'FIRST', 'LAST', 'AS', 'FOR', 'UPDATE',
  'TYPEOF', 'WHEN', 'THEN', 'ELSE', 'END', 'WITH', 'SECURITY_ENFORCED',
])

const functions = new Set([
  'COUNT', 'COUNT_DISTINCT', 'SUM', 'AVG', 'MIN', 'MAX',
  'COALESCE', 'NULLIF', 'CONCAT', 'UPPER', 'LOWER', 'TRIM',
  'LENGTH', 'LEN', 'SUBSTRING', 'SUBSTR', 'ABS', 'ROUND',
  'FLOOR', 'CEIL', 'CEILING',
])

const dateLiterals = new Set([
  'TODAY', 'YESTERDAY', 'TOMORROW', 'THIS_WEEK', 'LAST_WEEK', 'NEXT_WEEK',
  'THIS_MONTH', 'LAST_MONTH', 'NEXT_MONTH', 'THIS_QUARTER', 'LAST_QUARTER',
  'NEXT_QUARTER', 'THIS_YEAR', 'LAST_YEAR', 'NEXT_YEAR',
  'THIS_FISCAL_QUARTER', 'LAST_FISCAL_QUARTER', 'NEXT_FISCAL_QUARTER',
  'THIS_FISCAL_YEAR', 'LAST_FISCAL_YEAR', 'NEXT_FISCAL_YEAR',
])

const soqlParser: StreamParser<SoqlState> = {
  startState(): SoqlState {
    return { inString: false, inComment: false }
  },

  token(stream, state): string | null {
    // Handle strings
    if (state.inString) {
      while (!stream.eol()) {
        const ch = stream.next()
        if (ch === '\\') {
          stream.next()
        } else if (ch === "'") {
          state.inString = false
          break
        }
      }
      return t.string.toString()
    }

    // Skip whitespace
    if (stream.eatSpace()) return null

    // Line comments
    if (stream.match('--')) {
      stream.skipToEnd()
      return t.lineComment.toString()
    }

    const ch = stream.peek()

    // String start (single-quoted)
    if (ch === "'") {
      stream.next()
      state.inString = true
      while (!stream.eol()) {
        const c = stream.next()
        if (c === '\\') {
          stream.next()
        } else if (c === "'") {
          state.inString = false
          break
        }
      }
      return t.string.toString()
    }

    // Parameters (:paramName)
    if (ch === ':') {
      stream.next()
      if (stream.match(/^\w+/)) {
        return t.variableName.toString()
      }
      return t.punctuation.toString()
    }

    // Numbers (including dates like 2024-01-15)
    if (/\d/.test(ch!)) {
      // Try date format first: YYYY-MM-DD
      if (stream.match(/^\d{4}-\d{2}-\d{2}/)) {
        return t.number.toString()
      }
      stream.match(/^\d+(\.\d+)?/)
      return t.number.toString()
    }

    // Operators
    if (stream.match(/^(<>|!=|<=|>=|[=<>])/)) {
      return t.operator.toString()
    }

    // Punctuation
    if (/[(),.*]/.test(ch!)) {
      stream.next()
      return t.punctuation.toString()
    }

    // Words (identifiers, keywords, functions)
    if (stream.match(/^[a-zA-Z_]\w*/)) {
      const word = stream.current()
      const upper = word.toUpperCase()

      // Boolean/null
      if (upper === 'TRUE' || upper === 'FALSE' || upper === 'NULL') {
        return t.bool.toString()
      }

      // Keywords
      if (keywords.has(upper)) {
        return t.keyword.toString()
      }

      // Functions
      if (functions.has(upper)) {
        return t.function(t.variableName).toString()
      }

      // Date literals (exact or with N suffix like LAST_N_DAYS)
      if (dateLiterals.has(upper)) {
        return t.atom.toString()
      }

      // Dynamic date literals: LAST_N_DAYS:30, NEXT_N_MONTHS:6, etc.
      if (/^(LAST|NEXT)_N_(DAYS|WEEKS|MONTHS|QUARTERS|YEARS|FISCAL_QUARTERS|FISCAL_YEARS)$/i.test(word)) {
        // Consume the :N part if present
        if (stream.peek() === ':') {
          stream.next()
          stream.match(/^\d+/)
        }
        return t.atom.toString()
      }

      return t.variableName.toString()
    }

    // Fallback: consume a character
    stream.next()
    return null
  },
}

export const soqlLanguage = StreamLanguage.define(soqlParser)
