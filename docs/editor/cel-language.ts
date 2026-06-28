/**
 * CEL Expression Editor — Syntax Highlighting for CodeMirror 6
 *
 * A StreamLanguage parser that provides syntax highlighting for CEL expressions.
 * Highlights: strings, comments, numbers, operators, keywords, context variables,
 * built-in functions, and the fn.* custom function namespace.
 *
 * Dependencies:
 *   npm install @codemirror/language @lezer/highlight
 *
 * Usage:
 *   import { celLanguage } from './cel-language'
 *   // Pass as `language` extension to CodeMirror EditorState.create({ extensions: [celLanguage] })
 */

import { StreamLanguage, type StreamParser } from '@codemirror/language'
import { tags as t } from '@lezer/highlight'

interface CelState {
  inString: false | '"' | "'"
  inComment: boolean
}

const celParser: StreamParser<CelState> = {
  startState(): CelState {
    return { inString: false, inComment: false }
  },

  token(stream, state): string | null {
    // Handle strings (continuation from previous line — shouldn't happen in CEL, but safe)
    if (state.inString) {
      const quote = state.inString
      while (!stream.eol()) {
        const ch = stream.next()
        if (ch === '\\') {
          stream.next() // skip escaped char
        } else if (ch === quote) {
          state.inString = false
          break
        }
      }
      return t.string.toString()
    }

    // Skip whitespace
    if (stream.eatSpace()) return null

    // Line comments (//)
    if (stream.match('//')) {
      stream.skipToEnd()
      return t.lineComment.toString()
    }

    const ch = stream.peek()

    // String start (single or double quoted)
    if (ch === '"' || ch === "'") {
      stream.next()
      state.inString = ch as '"' | "'"
      while (!stream.eol()) {
        const c = stream.next()
        if (c === '\\') {
          stream.next()
        } else if (c === ch) {
          state.inString = false
          break
        }
      }
      return t.string.toString()
    }

    // Numbers (integers, floats, scientific notation)
    if (/\d/.test(ch!)) {
      stream.match(/^\d+(\.\d+)?([eE][+-]?\d+)?/)
      return t.number.toString()
    }

    // Multi-char operators
    if (stream.match(/^(&&|\|\||==|!=|>=|<=|[+\-*/%<>!])/)) {
      return t.operator.toString()
    }

    // Punctuation
    if (/[()[\]{},.:?]/.test(ch!)) {
      stream.next()
      return t.punctuation.toString()
    }

    // Words (identifiers, keywords, functions)
    if (stream.match(/^[a-zA-Z_]\w*/)) {
      const word = stream.current()

      // Boolean literals and null
      if (['true', 'false', 'null'].includes(word)) {
        return t.bool.toString()
      }

      // Keywords
      if (['in', 'has'].includes(word)) {
        return t.keyword.toString()
      }

      // fn.* namespace (custom functions)
      if (word === 'fn' && stream.peek() === '.') {
        return t.namespace.toString()
      }

      // Context variables
      if (['record', 'old', 'user', 'now'].includes(word)) {
        return t.variableName.toString()
      }

      // Built-in functions
      if (
        [
          'size',
          'contains',
          'startsWith',
          'endsWith',
          'matches',
          'int',
          'uint',
          'double',
          'string',
          'bool',
          'type',
          'duration',
          'timestamp',
        ].includes(word)
      ) {
        return t.function(t.variableName).toString()
      }

      // Default: identifier
      return t.variableName.toString()
    }

    // Fallback: consume a character
    stream.next()
    return null
  },
}

export const celLanguage = StreamLanguage.define(celParser)
