import { Environment } from '@marcbachmann/cel-js'
import type { Function } from '@/types/functions'

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

class FnNamespace {}

export function createCelEnvironment(functions: Function[]): Environment {
  const env = new Environment({ unlistedVariablesAreDyn: true })

  if (functions.length === 0) return env

  env.registerType('FnNamespace', FnNamespace)

  for (const fn of functions) {
    if (!fn.name || !fn.body) continue

    const params = fn.params ?? []
    // Receiver (FnNamespace) + dyn params
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
