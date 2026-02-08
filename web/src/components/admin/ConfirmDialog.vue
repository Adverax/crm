<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  title: string
  description: string
  confirmLabel?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: []
}>()

function onCancel() {
  emit('update:open', false)
}

function onConfirm() {
  emit('confirm')
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ props.title }}</DialogTitle>
        <DialogDescription>{{ props.description }}</DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" :disabled="props.loading" @click="onCancel">
          Отмена
        </Button>
        <Button variant="destructive" :disabled="props.loading" @click="onConfirm">
          {{ props.confirmLabel ?? 'Удалить' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
