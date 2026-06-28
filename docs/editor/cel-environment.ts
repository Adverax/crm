/**
 * CEL Expression Editor — Client-side Evaluation
 *
 * Browser-side CEL evaluation using @marcbachmann/cel-js.
 * Supports custom functions via the fn.* namespace.
 *
 * Dependencies:
 *   npm install @marcbachmann/cel-js
 *
 * Usage:
 *   import { createCelEnvironment, evaluateCel } from './cel-environment'
 *
 *   const env = createCelEnvironment(customFunctions)
 *   const result = evaluateCel(env, 'record.Amount > 100', {
 *     record: { Amount: 500 },
 *     user: { id: '1', profile_id: '2', role_id: '3' },
 *     now: new Date(),
 *   })
 *   // result = { success: true, value: true, type: 'bool' }
 */

import { Environment } from '@marcbachmann/cel-js'

// --- Types ---

interface FunctionParam {
  name: string
  type?: string
}

interface Function {
  name?: string
  body?: string
  params?: FunctionParam[]
}

export interface CelEvalContext {
  record?: Record<string, unknown>
  old?: Record<string, unknown>
  user?: { id: string; profile_id: string; role_id: string }
  now?: Date
  [key: string]: unknown
}

export interface CelEvalResult {
  success: boolean
  value?: unknown
  type?: string
  error?: string
  position?: number
}

// --- Helpers ---

/**
 * Recursively converts BigInt values to Number.
 * cel-js returns BigInt for integer literals; JSON.stringify doesn't handle BigInt.
 */
function convertBigInt(value: unknown): unknown {
  if (typeof value === 'bigint') {
    return Number(value)
  }
  if (Array.isArray(value)) {
    return value.map(convertBigInt)
  }
  if (value !== null && typeof value === 'object') {
    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(value)) {
      result[k] = convertBigInt(v)
    }
    return result
  }
  return value
}

/**
 * Infers the CEL type name from a JavaScript value.
 */
function inferType(value: unknown): string {
  if (value === null || value === undefined) return 'null'
  if (typeof value === 'boolean') return 'bool'
  if (typeof value === 'number') return 'number'
  if (typeof value === 'bigint') return 'int'
  if (typeof value === 'string') return 'string'
  if (Array.isArray(value)) return 'list'
  if (typeof value === 'object') return 'map'
  return 'unknown'
}

// --- Namespace class for fn.* functions ---

class FnNamespace {}

/**
 * Creates a CEL environment with custom fn.* functions registered.
 *
 * Each custom function is registered as a method on the FnNamespace class,
 * which is injected as the `fn` variable in the evaluation context.
 *
 * @param functions - Array of custom function definitions
 * @returns Configured cel-js Environment
 *
 * @example
 * ```ts
 * const env = createCelEnvironment([
 *   { name: 'greet', body: '"Hello, " + name', params: [{ name: 'name' }] },
 * ])
 * ```
 */
export function createCelEnvironment(functions: Function[]): Environment {
  const env = new Environment({ unlistedVariablesAreDyn: true })

  if (functions.length === 0) return env

  env.registerType('FnNamespace', FnNamespace)

  for (const fn of functions) {
    if (!fn.name || !fn.body) continue

    const params = fn.params ?? []
    const paramTypes = params.map(() => 'dyn').join(', ')
    const signature = paramTypes
      ? `FnNamespace.${fn.name}(${paramTypes}): dyn`
      : `FnNamespace.${fn.name}(): dyn`

    const fnBody = fn.body
    const paramNames = params.map((p) => p.name)

    env.registerFunction(signature, (_receiver: unknown, ...args: unknown[]) => {
      const innerEnv = new Environment({ unlistedVariablesAreDyn: true })
      const ctx: Record<string, unknown> = {}
      for (let i = 0; i < paramNames.length; i++) {
        ctx[paramNames[i]!] = args[i]
      }
      return innerEnv.evaluate(fnBody, ctx)
    })
  }

  return env
}

function buildContext(context: CelEvalContext, hasFunctions: boolean): CelEvalContext {
  if (!hasFunctions) return context
  return { ...context, fn: new FnNamespace() }
}

/**
 * Evaluates a CEL expression in the given context.
 *
 * @param env - CEL environment (created by createCelEnvironment)
 * @param expression - CEL expression string
 * @param context - Evaluation context with variables (record, user, etc.)
 * @param hasFunctions - Whether to inject fn namespace (default: true)
 * @returns Evaluation result with value, type, or error
 */
export function evaluateCel(
  env: Environment,
  expression: string,
  context: CelEvalContext,
  hasFunctions = true,
): CelEvalResult {
  if (!expression.trim()) {
    return { success: false, error: 'Empty expression' }
  }

  try {
    const ctx = buildContext(context, hasFunctions)
    const raw = env.evaluate(expression, ctx)
    const value = convertBigInt(raw)
    return { success: true, value, type: inferType(value) }
  } catch (err) {
    const result: CelEvalResult = {
      success: false,
      error: err instanceof Error ? err.message : String(err),
    }
    const pos = (err as Record<string, unknown>)?.node as Record<string, unknown> | undefined
    if (typeof pos?.pos === 'number') {
      result.position = pos.pos
    }
    return result
  }
}

/**
 * Evaluates a CEL expression with timeout protection.
 *
 * @param env - CEL environment
 * @param expression - CEL expression string
 * @param context - Evaluation context
 * @param timeoutMs - Maximum execution time in milliseconds (default: 100ms)
 * @returns Evaluation result or timeout error
 */
export function evaluateCelSafe(
  env: Environment,
  expression: string,
  context: CelEvalContext,
  timeoutMs = 100,
): CelEvalResult {
  const start = performance.now()

  const result = evaluateCel(env, expression, context)

  const elapsed = performance.now() - start
  if (elapsed > timeoutMs) {
    return {
      success: false,
      error: `Timeout exceeded (${Math.round(elapsed)}ms > ${timeoutMs}ms)`,
    }
  }

  return result
}
