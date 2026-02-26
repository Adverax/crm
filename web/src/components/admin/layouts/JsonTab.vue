<script setup lang="ts">
import { ref, watch } from 'vue'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import type { LayoutConfig } from '@/types/layouts'

const props = defineProps<{
  config: LayoutConfig
}>()

const emit = defineEmits<{
  'update:config': [config: LayoutConfig]
}>()

const jsonText = ref('')
const parseError = ref<string | null>(null)
// Guard to skip watch when we just emitted an update from user typing
let skipNextWatch = false

watch(
  () => props.config,
  (cfg) => {
    if (skipNextWatch) {
      skipNextWatch = false
      return
    }
    jsonText.value = JSON.stringify(cfg, null, 2)
    parseError.value = null
  },
  { immediate: true, deep: true },
)

function onInput(value: string | number) {
  jsonText.value = String(value)
  parseError.value = null
  try {
    const parsed = JSON.parse(jsonText.value) as LayoutConfig
    skipNextWatch = true
    emit('update:config', parsed)
  } catch (err) {
    parseError.value = err instanceof Error ? err.message : 'Invalid JSON'
  }
}
</script>

<template>
  <Card>
    <CardContent class="pt-6 space-y-2">
      <Label for="config-json">Config (JSON)</Label>
      <Textarea
        id="config-json"
        :model-value="jsonText"
        rows="24"
        class="font-mono text-sm"
        data-testid="json-config"
        @update:model-value="onInput"
      />
      <p v-if="parseError" class="text-sm text-destructive" data-testid="json-error">
        {{ parseError }}
      </p>
    </CardContent>
  </Card>
</template>
