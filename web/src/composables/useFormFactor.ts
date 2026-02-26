import { ref, onMounted, onUnmounted } from 'vue'

export type FormFactor = 'mobile' | 'tablet' | 'desktop'

function detectFormFactor(): FormFactor {
  const width = window.innerWidth
  if (width < 768) return 'mobile'
  if (width < 1024) return 'tablet'
  return 'desktop'
}

export function useFormFactor() {
  const formFactor = ref<FormFactor>(detectFormFactor())

  function onResize() {
    formFactor.value = detectFormFactor()
  }

  onMounted(() => window.addEventListener('resize', onResize))
  onUnmounted(() => window.removeEventListener('resize', onResize))

  return formFactor
}
