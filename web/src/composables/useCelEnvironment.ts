import { shallowRef, watch } from 'vue'
import type { Environment } from '@marcbachmann/cel-js'
import { useFunctionsStore } from '@/stores/functions'
import { createCelEnvironment, evaluateCel } from '@/lib/cel'
import type { CelEvalContext, CelEvalResult } from '@/lib/cel'

export function useCelEnvironment() {
  const store = useFunctionsStore()
  const env = shallowRef<Environment | null>(null)

  async function init() {
    await store.ensureLoaded()
    env.value = createCelEnvironment(store.functions)
  }

  watch(
    () => store.functions,
    (fns) => {
      if (fns.length > 0 || store.loaded) {
        env.value = createCelEnvironment(fns)
      }
    },
    { deep: true },
  )

  function evaluate(expression: string, context: CelEvalContext): CelEvalResult {
    if (!env.value) {
      return { success: false, error: 'CEL environment is not initialized' }
    }
    return evaluateCel(env.value, expression, context)
  }

  return { env, init, evaluate }
}
