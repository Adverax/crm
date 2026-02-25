<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { automationRulesApi } from '@/api/automation-rules'
import { http } from '@/api/http'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ActiveStatusBadge from '@/components/admin/ActiveStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import type { AutomationRule, EventType, ExecutionMode } from '@/types/automation-rules'

const props = defineProps<{
  ruleId: string
}>()

const router = useRouter()
const toast = useToast()

const rule = ref<AutomationRule | null>(null)
const objectLabel = ref('')
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)

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

async function loadRule() {
  loading.value = true
  try {
    const response = await automationRulesApi.get(props.ruleId)
    rule.value = response.data
    try {
      const objResp = await http.get<{ data: { label: string } }>(
        `/api/v1/admin/metadata/objects/${response.data.objectId}`,
      )
      objectLabel.value = objResp.data.label
    } catch {
      // object label won't be available for breadcrumbs
    }
    form.value = {
      name: response.data.name,
      description: response.data.description ?? '',
      eventType: response.data.eventType,
      condition: response.data.condition ?? '',
      procedureCode: response.data.procedureCode,
      executionMode: response.data.executionMode,
      sortOrder: response.data.sortOrder,
      isActive: response.data.isActive,
    }
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadRule)
watch(() => props.ruleId, loadRule)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onEventTypeChange(val: any) {
  form.value.eventType = String(val) as EventType
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onExecutionModeChange(val: any) {
  form.value.executionMode = String(val) as ExecutionMode
}

async function onSave() {
  submitting.value = true
  try {
    await automationRulesApi.update(props.ruleId, {
      name: form.value.name,
      description: form.value.description,
      event_type: form.value.eventType,
      condition: form.value.condition || null,
      procedure_code: form.value.procedureCode,
      execution_mode: form.value.executionMode,
      sort_order: form.value.sortOrder,
      is_active: form.value.isActive,
    })
    toast.success('Automation rule updated')
    await loadRule()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await automationRulesApi.delete(props.ruleId)
    toast.success('Automation rule deleted')
    router.push({ name: 'admin-automation-rules' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-automation-rules' })
}

function eventLabel(type: string): string {
  return type.replace(/_/g, ' ')
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Automation Rules', to: '/admin/metadata/automation-rules' },
  { label: objectLabel.value || '...' },
  { label: rule.value?.name ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="loading && !rule" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="rule">
      <PageHeader :title="rule.name" :breadcrumbs="breadcrumbs">
        <template #actions>
          <div class="flex items-center gap-2">
            <ActiveStatusBadge :is-active="rule.isActive" />
            <Badge variant="outline">{{ eventLabel(rule.eventType) }}</Badge>
            <Badge variant="outline">{{ rule.procedureCode }}</Badge>
            <IconButton
              :icon="Trash2"
              tooltip="Delete rule"
              variant="destructive"
              data-testid="delete-rule-btn"
              @click="showDeleteDialog = true"
            />
          </div>
        </template>
      </PageHeader>

      <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSave">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="space-y-2">
              <Label for="name">Name</Label>
              <Input id="name" v-model="form.name" required data-testid="field-name" />
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
          <Button type="submit" :disabled="submitting" data-testid="save-btn">
            Save
          </Button>
          <IconButton :icon="X" tooltip="Cancel" variant="outline" data-testid="cancel-btn" @click="onCancel" />
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete automation rule?"
        :description="`Rule '${rule.name}' will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
