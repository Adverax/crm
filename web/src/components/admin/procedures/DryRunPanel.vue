<script setup lang="ts">
import { ref } from 'vue'
import { proceduresApi } from '@/api/procedures'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Play } from 'lucide-vue-next'
import type { ExecutionResult } from '@/types/procedures'

const props = defineProps<{
  procedureId: string
}>()

const toast = useToast()
const inputJson = ref('{}')
const running = ref(false)
const result = ref<ExecutionResult | null>(null)
const error = ref<string | null>(null)

async function onRun() {
  running.value = true
  result.value = null
  error.value = null

  let input: Record<string, unknown> = {}
  try {
    input = JSON.parse(inputJson.value)
  } catch {
    error.value = 'Invalid JSON input'
    running.value = false
    return
  }

  try {
    const response = await proceduresApi.dryRun(props.procedureId, { input })
    result.value = response.data
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    error.value = msg
    toast.errorFromApi(err)
  } finally {
    running.value = false
  }
}
</script>

<template>
  <div class="grid grid-cols-2 gap-4">
    <!-- Input -->
    <Card>
      <CardContent class="pt-6 space-y-4">
        <Label>Input JSON</Label>
        <Textarea
          v-model="inputJson"
          rows="10"
          class="font-mono text-sm"
          data-testid="dry-run-input"
        />
        <Button :disabled="running" data-testid="dry-run-btn" @click="onRun">
          <Play class="h-4 w-4 mr-2" />
          {{ running ? 'Running...' : 'Run' }}
        </Button>
      </CardContent>
    </Card>

    <!-- Output -->
    <Card>
      <CardContent class="pt-6 space-y-4">
        <Label>Result</Label>

        <div v-if="error" class="text-sm text-destructive" data-testid="dry-run-error">
          {{ error }}
        </div>

        <template v-else-if="result">
          <div class="flex items-center gap-2 mb-2">
            <Badge :variant="result.success ? 'default' : 'destructive'">
              {{ result.success ? 'Success' : 'Failed' }}
            </Badge>
          </div>

          <div v-if="result.result" class="space-y-1">
            <Label class="text-xs">Output</Label>
            <pre class="text-xs font-mono bg-muted p-2 rounded overflow-auto max-h-40" data-testid="dry-run-result">{{ JSON.stringify(result.result, null, 2) }}</pre>
          </div>

          <div v-if="result.warnings && result.warnings.length > 0" class="space-y-1">
            <Label class="text-xs">Warnings</Label>
            <div v-for="(w, i) in result.warnings" :key="i" class="text-xs text-yellow-600">
              {{ w.command }}: {{ w.message }}
            </div>
          </div>

          <div v-if="result.trace && result.trace.length > 0" class="space-y-1">
            <Label class="text-xs">Trace</Label>
            <div v-for="(t, i) in result.trace" :key="i" class="flex items-center gap-2 text-xs">
              <Badge
                :variant="t.status === 'ok' ? 'default' : t.status === 'skipped' ? 'outline' : 'destructive'"
                class="text-[10px] px-1 py-0"
              >
                {{ t.status }}
              </Badge>
              <span class="font-mono">{{ t.step }}</span>
              <span class="text-muted-foreground">{{ t.durationMs }}ms</span>
            </div>
          </div>
        </template>

        <div v-else class="text-sm text-muted-foreground">
          Run the procedure to see results.
        </div>
      </CardContent>
    </Card>
  </div>
</template>
