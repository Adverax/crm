function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
}

function toCamelCase(str: string): string {
  return str.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase())
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value) && !(value instanceof Date)
}

function convertKeys(obj: unknown, converter: (key: string) => string): unknown {
  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeys(item, converter))
  }
  if (isPlainObject(obj)) {
    const result: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(obj)) {
      result[converter(key)] = convertKeys(value, converter)
    }
    return result
  }
  return obj
}

export function snakeToCamel<T>(obj: unknown): T {
  return convertKeys(obj, toCamelCase) as T
}

export function camelToSnake<T>(obj: unknown): T {
  return convertKeys(obj, toSnakeCase) as T
}
