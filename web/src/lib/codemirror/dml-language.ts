import { StreamLanguage, type StreamParser } from '@codemirror/language'
import { tags as t } from '@lezer/highlight'

interface DmlState {
  inString: boolean
}

const keywords = new Set([
  'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE', 'FROM',
  'WHERE', 'UPSERT', 'ON', 'AND', 'OR', 'NOT', 'IN', 'LIKE',
  'IS', 'BETWEEN', 'EXISTS',
])

const functions = new Set([
  'COALESCE', 'NULLIF', 'CONCAT', 'UPPER', 'LOWER', 'TRIM',
  'LENGTH', 'LEN', 'SUBSTRING', 'SUBSTR', 'ABS', 'ROUND',
  'FLOOR', 'CEIL', 'CEILING',
])

const dmlParser: StreamParser<DmlState> = {
  startState(): DmlState {
    return { inString: false }
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

      return t.variableName.toString()
    }

    // Fallback: consume a character
    stream.next()
    return null
  },
}

export const dmlLanguage = StreamLanguage.define(dmlParser)
