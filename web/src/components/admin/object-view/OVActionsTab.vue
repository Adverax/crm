<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { OVAction } from '@/types/object-views'

const props = defineProps<{
  actions: OVAction[]
}>()

const emit = defineEmits<{
  'update:actions': [value: OVAction[]]
}>()

function addAction() {
  const updated = [...props.actions, {
    key: `action_${Date.now()}`,
    label: 'New Action',
    type: 'secondary',
    icon: '',
    visibilityExpr: '',
  }]
  emit('update:actions', updated)
}

function removeAction(index: number) {
  const updated = [...props.actions]
  updated.splice(index, 1)
  emit('update:actions', updated)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onActionTypeChange(index: number, value: any) {
  const action = props.actions[index]
  if (action) action.type = String(value)
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Actions</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add action"
        variant="outline"
        data-testid="add-action-btn"
        @click="addAction"
      />
    </div>

    <Card
      v-for="(action, aIdx) in actions"
      :key="aIdx"
      data-testid="action-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Key</Label>
            <Input v-model="action.key" placeholder="action_key" class="font-mono" />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Label</Label>
            <Input v-model="action.label" placeholder="Action Label" />
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">Type</Label>
              <Select
                :model-value="action.type"
                @update:model-value="(v) => onActionTypeChange(aIdx, v)"
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="primary">Primary</SelectItem>
                  <SelectItem value="secondary">Secondary</SelectItem>
                  <SelectItem value="danger">Danger</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete action"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeAction(aIdx)"
            />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Icon (lucide name)</Label>
            <Input v-model="action.icon" placeholder="mail" class="font-mono" />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Visibility Expression (CEL)</Label>
            <Input
              v-model="action.visibilityExpr"
              placeholder="record.status == 'draft'"
              class="font-mono"
            />
          </div>
        </div>
      </CardContent>
    </Card>

    <div v-if="actions.length === 0" class="text-sm text-muted-foreground">
      No custom actions configured.
    </div>
  </div>
</template>
