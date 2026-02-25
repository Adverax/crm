<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { automationRulesApi } from '@/api/automation-rules'
import { http } from '@/api/http'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
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
import { Switch } from '@/components/ui/switch'
import type { EventType, ExecutionMode } from '@/types/automation-rules'

const router = useRouter()
const route = useRoute()
const toast = useToast()
const submitting = ref(false)

const objectId = computed(() => String(route.params.objectId ?? ''))
const objectLabel = ref('')

onMounted(async () => {
  try {
    const resp = await http.get<{ data: { label: string } }>(
      `/api/v1/admin/metadata/objects/${objectId.value}`,
    )
    objectLabel.value = resp.data.label
  } catch {
    // object label won't be available for breadcrumbs
  }
})

const form = ref({
  name: '',
  description: '',
  eventType: 'after_insert' as EventType,
  condition: '',
  procedureCode: '',
  executionMode: 'per_record' as ExecutionMode,
  sortOrder: 0,
  isActive: true,
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onEventTypeChange(val: any) {
  form.value.eventType = String(val) as EventType
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onExecutionModeChange(val: any) {
  form.value.executionMode = String(val) as ExecutionMode
}

async function onSubmit() {
  submitting.value = true
  try {
    const response = await automationRulesApi.create(objectId.value, {
      name: form.value.name,
      description: form.value.description || undefined,
      event_type: form.value.eventType,
      condition: form.value.condition || null,
      procedure_code: form.value.procedureCode,
      execution_mode: form.value.executionMode,
      sort_order: form.value.sortOrder,
      is_active: form.value.isActive,
    })
    toast.success('Automation rule created')
    router.push({ name: 'admin-automation-rule-detail', params: { ruleId: response.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-automation-rules' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Automation Rules', to: '/admin/metadata/automation-rules' },
  { label: objectLabel.value || '...' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Automation Rule" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="space-y-2">
            <Label for="name">Name</Label>
            <Input id="name" v-model="form.name" required placeholder="Rule name" data-testid="field-name" />
          </div>
          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea id="description" v-model="form.description" rows="2" data-testid="field-description" />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label>Event Type</Label>
              <Select :model-value="form.eventType" @update:model-value="onEventTypeChange">
                <SelectTrigger data-testid="field-event-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="before_insert">Before Insert</SelectItem>
                  <SelectItem value="after_insert">After Insert</SelectItem>
                  <SelectItem value="before_update">Before Update</SelectItem>
                  <SelectItem value="after_update">After Update</SelectItem>
                  <SelectItem value="before_delete">Before Delete</SelectItem>
                  <SelectItem value="after_delete">After Delete</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="procedure-code">Procedure Code</Label>
              <Input
                id="procedure-code"
                v-model="form.procedureCode"
                required
                placeholder="notify_manager"
                class="font-mono"
                data-testid="field-procedure-code"
              />
            </div>
          </div>
          <div class="space-y-2">
            <Label for="condition">Condition (CEL expression, optional)</Label>
            <Input
              id="condition"
              v-model="form.condition"
              placeholder="new.Status == 'Approved'"
              class="font-mono"
              data-testid="field-condition"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label>Execution Mode</Label>
              <Select :model-value="form.executionMode" @update:model-value="onExecutionModeChange">
                <SelectTrigger data-testid="field-execution-mode">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="per_record">Per Record</SelectItem>
                  <SelectItem value="per_batch">Per Batch</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="sort-order">Sort Order</Label>
              <Input
                id="sort-order"
                v-model.number="form.sortOrder"
                type="number"
                data-testid="field-sort-order"
              />
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Switch
              :checked="form.isActive"
              data-testid="field-is-active"
              @update:checked="form.isActive = $event"
            />
            <Label>Active</Label>
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton :icon="X" tooltip="Cancel" variant="outline" data-testid="cancel-btn" @click="onCancel" />
      </div>
    </form>
  </div>
</template>
