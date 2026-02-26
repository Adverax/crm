<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import type { ObjectView } from '@/types/object-views'

defineProps<{
  view: ObjectView
  form: { label: string; description: string }
}>()

const emit = defineEmits<{
  'update:label': [value: string]
  'update:description': [value: string]
}>()
</script>

<template>
  <Card>
    <CardContent class="pt-6 space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <Label>API Name</Label>
          <Input :model-value="view.apiName" disabled class="font-mono" />
        </div>
        <div class="space-y-2">
          <Label for="label">Label</Label>
          <Input
            id="label"
            :model-value="form.label"
            required
            data-testid="field-label"
            @update:model-value="(v: string | number) => emit('update:label', String(v))"
          />
        </div>
      </div>

      <div class="space-y-2">
        <Label for="description">Description</Label>
        <Textarea
          id="description"
          :model-value="form.description"
          rows="2"
          data-testid="field-description"
          @update:model-value="(v: string | number) => emit('update:description', String(v))"
        />
      </div>

      <div class="space-y-2">
        <Label>Profile</Label>
        <Input :model-value="view.profileId ?? 'None (global)'" disabled class="font-mono text-xs" />
      </div>
    </CardContent>
  </Card>
</template>
