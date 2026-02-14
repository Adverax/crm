// Converts snake_case string literal to camelCase
type SnakeToCamel<S extends string> =
  S extends `${infer P}_${infer R}`
    ? `${P}${Capitalize<SnakeToCamel<R>>}`
    : S

// Recursively converts all keys of an object from snake_case to camelCase
export type CamelCaseKeys<T> =
  T extends Array<infer U>
    ? CamelCaseKeys<U>[]
    : T extends object
      ? { [K in keyof T as K extends string ? SnakeToCamel<K> : K]: CamelCaseKeys<T[K]> }
      : T
