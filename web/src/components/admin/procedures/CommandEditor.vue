<script setup lang="ts">
import { computed } from 'vue'
import { IconButton } from '@/components/ui/icon-button'
import { ChevronUp, ChevronDown, Trash2 } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import KeyValueEditor from './KeyValueEditor.vue'
import type { CommandDef } from '@/types/procedures'

const props = defineProps<{
  command: CommandDef
  index: number
  total: number
}>()

const emit = defineEmits<{
  update: [cmd: CommandDef]
  remove: []
  moveUp: []
  moveDown: []
}>()

function update(partial: Partial<CommandDef>) {
  emit('update', { ...props.command, ...partial })
}

const category = computed(() => {
  const parts = props.command.type.split('.')
  return parts[0] ?? 'unknown'
})

const subtype = computed(() => {
  const parts = props.command.type.split('.')
  return parts[1] ?? ''
})

const categoryColors: Record<string, string> = {
  record: 'bg-blue-100 text-blue-800',
  compute: 'bg-green-100 text-green-800',
  flow: 'bg-purple-100 text-purple-800',
  integration: 'bg-orange-100 text-orange-800',
  notification: 'bg-yellow-100 text-yellow-800',
  wait: 'bg-gray-100 text-gray-800',
}

const colorClass = computed(() => categoryColors[category.value] ?? 'bg-gray-100 text-gray-800')

const isRecord = computed(() => category.value === 'record')
const isCompute = computed(() => category.value === 'compute')
const isFlow = computed(() => category.value === 'flow')
const isIntegration = computed(() => category.value === 'integration')
</script>

<template>
  <Card class="border-l-4" :class="colorClass.replace('bg-', 'border-l-').replace(/\s.*/, '')" data-testid="command-card">
    <CardContent class="py-3 space-y-3">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Badge :class="colorClass" variant="secondary">{{ command.type }}</Badge>
          <span class="text-xs text-muted-foreground">#{{ index + 1 }}</span>
        </div>
        <div class="flex items-center gap-1">
          <IconButton
            :icon="ChevronUp"
            tooltip="Move up"
            variant="ghost"
            size="icon-sm"
            :disabled="index === 0"
            @click="$emit('moveUp')"
          />
          <IconButton
            :icon="ChevronDown"
            tooltip="Move down"
            variant="ghost"
            size="icon-sm"
            :disabled="index === total - 1"
            @click="$emit('moveDown')"
          />
          <IconButton
            :icon="Trash2"
            tooltip="Remove command"
            variant="ghost"
            size="icon-sm"
            class="text-destructive"
            data-testid="remove-command-btn"
            @click="$emit('remove')"
          />
        </div>
      </div>

      <!-- Common fields -->
      <div class="grid grid-cols-2 gap-2">
        <div class="space-y-1">
          <Label class="text-xs">Step name (as)</Label>
          <Input
            :model-value="command.as ?? ''"
            placeholder="step_name"
            class="font-mono h-8 text-sm"
            data-testid="field-as"
            @input="update({ as: ($event.target as HTMLInputElement).value || undefined })"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">When condition</Label>
          <Input
            :model-value="command.when ?? ''"
            placeholder="$.input.active == true"
            class="font-mono h-8 text-sm"
            @input="update({ when: ($event.target as HTMLInputElement).value || undefined })"
          />
        </div>
      </div>

      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <Switch
            :checked="command.optional ?? false"
            @update:checked="update({ optional: $event })"
          />
          <Label class="text-xs">Optional</Label>
        </div>
      </div>

      <!-- Record fields -->
      <template v-if="isRecord">
        <div v-if="subtype === 'query'" class="space-y-1">
          <Label class="text-xs">SOQL Query</Label>
          <Textarea
            :model-value="command.query ?? ''"
            placeholder="SELECT Id, Name FROM Account WHERE ..."
            rows="2"
            class="font-mono text-sm"
            @input="update({ query: ($event.target as HTMLTextAreaElement).value })"
          />
        </div>
        <template v-else>
          <div class="grid grid-cols-2 gap-2">
            <div class="space-y-1">
              <Label class="text-xs">Object</Label>
              <Input
                :model-value="command.object ?? ''"
                placeholder="Account"
                class="h-8 text-sm"
                @input="update({ object: ($event.target as HTMLInputElement).value })"
              />
            </div>
            <div v-if="subtype !== 'create'" class="space-y-1">
              <Label class="text-xs">Record ID</Label>
              <Input
                :model-value="command.id ?? ''"
                placeholder="$.input.record_id"
                class="font-mono h-8 text-sm"
                @input="update({ id: ($event.target as HTMLInputElement).value })"
              />
            </div>
          </div>
          <KeyValueEditor
            v-if="subtype === 'create' || subtype === 'update'"
            :model-value="command.data"
            label="Data"
            key-placeholder="FieldName"
            value-placeholder="$.input.value"
            @update:model-value="update({ data: $event })"
          />
        </template>
      </template>

      <!-- Compute fields -->
      <template v-if="isCompute && subtype === 'transform'">
        <KeyValueEditor
          :model-value="command.value"
          label="Value mapping"
          key-placeholder="variable_name"
          value-placeholder="$.input.field + '_suffix'"
          @update:model-value="update({ value: $event })"
        />
      </template>

      <template v-if="isCompute && subtype === 'validate'">
        <div class="space-y-1">
          <Label class="text-xs">Condition</Label>
          <Input
            :model-value="command.condition ?? ''"
            placeholder="$.input.email != ''"
            class="font-mono h-8 text-sm"
            @input="update({ condition: ($event.target as HTMLInputElement).value })"
          />
        </div>
        <div class="grid grid-cols-2 gap-2">
          <div class="space-y-1">
            <Label class="text-xs">Error Code</Label>
            <Input
              :model-value="command.code ?? ''"
              placeholder="validation_error"
              class="h-8 text-sm"
              @input="update({ code: ($event.target as HTMLInputElement).value })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Error Message</Label>
            <Input
              :model-value="command.message ?? ''"
              placeholder="Validation failed"
              class="h-8 text-sm"
              @input="update({ message: ($event.target as HTMLInputElement).value })"
            />
          </div>
        </div>
      </template>

      <template v-if="isCompute && subtype === 'fail'">
        <div class="grid grid-cols-2 gap-2">
          <div class="space-y-1">
            <Label class="text-xs">Error Code</Label>
            <Input
              :model-value="command.code ?? ''"
              class="h-8 text-sm"
              @input="update({ code: ($event.target as HTMLInputElement).value })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Error Message</Label>
            <Input
              :model-value="command.message ?? ''"
              class="h-8 text-sm"
              @input="update({ message: ($event.target as HTMLInputElement).value })"
            />
          </div>
        </div>
      </template>

      <!-- Flow fields -->
      <template v-if="isFlow && subtype === 'if'">
        <div class="space-y-1">
          <Label class="text-xs">Condition</Label>
          <Input
            :model-value="command.condition ?? ''"
            placeholder="$.input.amount > 1000"
            class="font-mono h-8 text-sm"
            @input="update({ condition: ($event.target as HTMLInputElement).value })"
          />
        </div>
      </template>

      <template v-if="isFlow && subtype === 'call'">
        <div class="space-y-1">
          <Label class="text-xs">Procedure Code</Label>
          <Input
            :model-value="command.procedure ?? ''"
            placeholder="other_procedure"
            class="font-mono h-8 text-sm"
            @input="update({ procedure: ($event.target as HTMLInputElement).value })"
          />
        </div>
        <KeyValueEditor
          :model-value="command.input"
          label="Input mapping"
          key-placeholder="param_name"
          value-placeholder="$.input.value"
          @update:model-value="update({ input: $event })"
        />
      </template>

      <template v-if="isFlow && subtype === 'try'">
        <div class="text-sm text-muted-foreground" data-testid="flow-try-info">
          Commands in 'try' block execute normally. If an error occurs, 'catch' block runs with <code class="text-xs bg-muted px-1 rounded">$.error</code> available.
        </div>
      </template>

      <!-- Integration fields -->
      <template v-if="isIntegration">
        <div class="grid grid-cols-3 gap-2">
          <div class="space-y-1">
            <Label class="text-xs">Credential</Label>
            <Input
              :model-value="command.credential ?? ''"
              placeholder="my_api"
              class="font-mono h-8 text-sm"
              @input="update({ credential: ($event.target as HTMLInputElement).value })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Method</Label>
            <Input
              :model-value="command.method ?? 'GET'"
              placeholder="GET"
              class="h-8 text-sm"
              @input="update({ method: ($event.target as HTMLInputElement).value })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Path</Label>
            <Input
              :model-value="command.path ?? ''"
              placeholder="/api/v1/data"
              class="font-mono h-8 text-sm"
              @input="update({ path: ($event.target as HTMLInputElement).value })"
            />
          </div>
        </div>
        <KeyValueEditor
          :model-value="command.headers"
          label="Headers"
          key-placeholder="Content-Type"
          value-placeholder="application/json"
          @update:model-value="update({ headers: $event })"
        />
        <div class="space-y-1">
          <Label class="text-xs">Body</Label>
          <Textarea
            :model-value="command.body ?? ''"
            placeholder='{"key": "$.input.value"}'
            rows="2"
            class="font-mono text-sm"
            @input="update({ body: ($event.target as HTMLTextAreaElement).value || undefined })"
          />
        </div>
      </template>

      <!-- Retry config (available for any command) -->
      <div class="space-y-2" data-testid="retry-section">
        <div class="flex items-center gap-2">
          <Switch
            :checked="!!command.retry"
            data-testid="retry-switch"
            @update:checked="update({ retry: $event ? { maxAttempts: 2, delayMs: 1000 } : undefined })"
          />
          <Label class="text-xs">Enable Retry</Label>
        </div>
        <div v-if="command.retry" class="grid grid-cols-3 gap-2" data-testid="retry-config">
          <div class="space-y-1">
            <Label class="text-xs">Max Attempts (1-5)</Label>
            <Input
              type="number"
              :model-value="command.retry.maxAttempts"
              min="1"
              max="5"
              class="h-8 text-sm"
              data-testid="retry-max-attempts"
              @input="update({ retry: { ...command.retry!, maxAttempts: parseInt(($event.target as HTMLInputElement).value) || 1 } })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Delay (ms)</Label>
            <Input
              type="number"
              :model-value="command.retry.delayMs"
              min="100"
              max="60000"
              class="h-8 text-sm"
              data-testid="retry-delay-ms"
              @input="update({ retry: { ...command.retry!, delayMs: parseInt(($event.target as HTMLInputElement).value) || 100 } })"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Backoff Mult</Label>
            <Input
              type="number"
              :model-value="command.retry.backoffMult ?? 1"
              min="1"
              max="10"
              class="h-8 text-sm"
              data-testid="retry-backoff-mult"
              @input="update({ retry: { ...command.retry!, backoffMult: parseInt(($event.target as HTMLInputElement).value) || 1 } })"
            />
          </div>
        </div>
      </div>

      <!-- Stub types -->
      <div v-if="category === 'notification' || category === 'wait'" class="text-sm text-muted-foreground italic">
        This command type is not yet implemented.
      </div>
    </CardContent>
  </Card>
</template>