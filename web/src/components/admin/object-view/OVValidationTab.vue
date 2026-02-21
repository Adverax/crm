<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { OVValidation } from '@/types/object-views'

const props = defineProps<{
  validation: OVValidation[]
}>()

const emit = defineEmits<{
  'update:validation': [value: OVValidation[]]
}>()

function addValidation() {
  const updated: OVValidation[] = [...props.validation, { expr: '', message: '', code: '', severity: 'error' as const, when: '' }]
  emit('update:validation', updated)
}

function removeValidation(index: number) {
  const updated = [...props.validation]
  updated.splice(index, 1)
  emit('update:validation', updated)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onSeverityChange(index: number, value: any) {
  const v = props.validation[index]
  if (v) v.severity = String(value) as OVValidation['severity']
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Validation</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add validation rule"
        variant="outline"
        data-testid="add-validation-btn"
        @click="addValidation"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      View-scoped validation rules (additive with metadata-level rules).
    </p>

    <Card
      v-for="(rule, idx) in validation"
      :key="idx"
      data-testid="validation-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Message</Label>
            <Input v-model="rule.message" placeholder="Amount must be positive" />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Code</Label>
            <Input v-model="rule.code" placeholder="invalid_amount" class="font-mono" />
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">Severity</Label>
              <Select
                :model-value="rule.severity"
                @update:model-value="(v) => onSeverityChange(idx, v)"
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="error">error</SelectItem>
                  <SelectItem value="warning">warning</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete validation rule"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeValidation(idx)"
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs">Expression (CEL)</Label>
          <Textarea
            v-model="rule.expr"
            placeholder="record.amount > 0"
            class="font-mono text-sm"
            rows="2"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">When (CEL)</Label>
          <Input v-model="rule.when" placeholder="has(record.amount)" class="font-mono" />
        </div>
      </CardContent>
    </Card>

    <div v-if="validation.length === 0" class="text-sm text-muted-foreground">
      No view-scoped validation rules.
    </div>
  </div>
</template>
