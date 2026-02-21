<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import type { OVMutation } from '@/types/object-views'

const props = defineProps<{
  mutations: OVMutation[]
}>()

const emit = defineEmits<{
  'update:mutations': [value: OVMutation[]]
}>()

function addMutation() {
  const updated: OVMutation[] = [...props.mutations, { dml: '', foreach: '', sync: undefined, when: '' }]
  emit('update:mutations', updated)
}

function removeMutation(index: number) {
  const updated = [...props.mutations]
  updated.splice(index, 1)
  emit('update:mutations', updated)
}

function ensureSync(index: number) {
  const mutation = props.mutations[index]
  if (mutation && !mutation.sync) {
    mutation.sync = { key: '', value: '' }
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Mutations</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add mutation"
        variant="outline"
        data-testid="add-mutation-btn"
        @click="addMutation"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      DML operations scoped to this Object View context.
    </p>

    <Card
      v-for="(mutation, idx) in mutations"
      :key="idx"
      data-testid="mutation-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Foreach (CEL)</Label>
            <Input v-model="mutation.foreach" placeholder="queries.line_items" class="font-mono" />
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">When (CEL)</Label>
              <Input v-model="mutation.when" placeholder="record.status == 'confirmed'" class="font-mono" />
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete mutation"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeMutation(idx)"
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs">DML</Label>
          <Textarea
            v-model="mutation.dml"
            placeholder="INSERT INTO LineItem (order_id, product_id) VALUES (:recordId, :item.product_id)"
            class="font-mono text-sm"
            rows="3"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">Sync (optional key/value mapping)</Label>
          <div class="grid grid-cols-2 gap-3">
            <Input
              :model-value="mutation.sync?.key ?? ''"
              placeholder="sync key"
              class="font-mono"
              @update:model-value="() => { ensureSync(idx); }"
              @input="(e: Event) => { ensureSync(idx); if (mutation.sync) mutation.sync.key = (e.target as HTMLInputElement).value }"
            />
            <Input
              :model-value="mutation.sync?.value ?? ''"
              placeholder="sync value"
              class="font-mono"
              @input="(e: Event) => { ensureSync(idx); if (mutation.sync) mutation.sync.value = (e.target as HTMLInputElement).value }"
            />
          </div>
        </div>
      </CardContent>
    </Card>

    <div v-if="mutations.length === 0" class="text-sm text-muted-foreground">
      No mutations configured. Mutations define write operations for this view.
    </div>
  </div>
</template>
