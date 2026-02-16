import { describe, it, expect } from 'vitest'
import { createCelEnvironment, evaluateCel, evaluateCelSafe } from '@/lib/cel'
import type { Function } from '@/types/functions'

describe('createCelEnvironment', () => {
  it('creates environment with empty function list', () => {
    const env = createCelEnvironment([])
    expect(env).toBeDefined()
  })

  it('registers custom functions', () => {
    const functions: Function[] = [
      {
        id: '1',
        name: 'double',
        body: 'x * 2',
        params: [{ name: 'x', type: 'number' }],
        returnType: 'number',
        createdAt: '',
        updatedAt: '',
      },
    ]
    const env = createCelEnvironment(functions)
    expect(env).toBeDefined()
  })
})

describe('evaluateCel', () => {
  it('evaluates arithmetic: 1 + 2 = 3', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, '1 + 2', {})
    expect(result.success).toBe(true)
    expect(result.value).toBe(3)
    expect(result.type).toBe('number')
  })

  it('evaluates with record context', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, 'record.Amount > 100', {
      record: { Amount: 200 },
    })
    expect(result.success).toBe(true)
    expect(result.value).toBe(true)
    expect(result.type).toBe('bool')
  })

  it('evaluates custom function call: fn.double(5) = 10', () => {
    const functions: Function[] = [
      {
        id: '1',
        name: 'double',
        body: 'x * 2',
        params: [{ name: 'x', type: 'number' }],
        returnType: 'number',
        createdAt: '',
        updatedAt: '',
      },
    ]
    const env = createCelEnvironment(functions)
    const result = evaluateCel(env, 'fn.double(5)', {})
    expect(result.success).toBe(true)
    expect(result.value).toBe(10)
  })

  it('returns error for invalid expression', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, '1 +', {})
    expect(result.success).toBe(false)
    expect(result.error).toBeDefined()
  })

  it('returns error for empty expression', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, '', {})
    expect(result.success).toBe(false)
    expect(result.error).toBe('Empty expression')
  })

  it('evaluates string operations', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, '"hello" + " " + "world"', {})
    expect(result.success).toBe(true)
    expect(result.value).toBe('hello world')
    expect(result.type).toBe('string')
  })

  it('evaluates boolean expression', () => {
    const env = createCelEnvironment([])
    const result = evaluateCel(env, 'true && false', {})
    expect(result.success).toBe(true)
    expect(result.value).toBe(false)
    expect(result.type).toBe('bool')
  })

  it('multiplies int literal by BigInt parameter: 2 * x', () => {
    const env = createCelEnvironment([])
    // BigInt(5) matches CEL int literal `2` → int * int = int
    const result = evaluateCel(env, '2 * x', { x: BigInt(5) }, false)
    expect(result.success).toBe(true)
    expect(result.value).toBe(10)
  })
})

describe('evaluateCelSafe', () => {
  it('returns result when within timeout', () => {
    const env = createCelEnvironment([])
    const result = evaluateCelSafe(env, '1 + 2', {}, 1000)
    expect(result.success).toBe(true)
    expect(result.value).toBe(3)
  })

  it('returns error for invalid expression', () => {
    const env = createCelEnvironment([])
    const result = evaluateCelSafe(env, 'invalid(((', {}, 1000)
    expect(result.success).toBe(false)
    expect(result.error).toBeDefined()
  })
})
