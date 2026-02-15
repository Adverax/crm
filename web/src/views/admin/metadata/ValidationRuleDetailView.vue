<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { validationRulesApi } from '@/api/validationRules'
import { useToast } from '@/composables/useToast'
import { http } from '@/api/http'
import PageHeader from '@/components/admin/PageHeader.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ValidationRule } from '@/types/validationRules'

const props = defineProps<{
  objectId: string
  ruleId: string
}>()

const objectApiName = ref('')

onMounted(async () => {
  try {
    const resp = await http.get<{ data: { api_name: string } }>(
      `/api/v1/admin/metadata/objects/${props.objectId}`,
    )
    objectApiName.value = resp.data.api_name
  } catch {
    // object api name won't be available for field picker
  }
})

const router = useRouter()
const toast = useToast()

const rule = ref<ValidationRule | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)

const form = ref({
  label: '',
  expression: '',
  errorMessage: '',
  errorCode: '',
  severity: 'error' as 'error' | 'warning',
  whenExpression: '',
  appliesTo: '',
  sortOrder: 0,
  isActive: true,
  description: '',
})

async function loadRule() {
  loading.value = true
  try {
    const response = await validationRulesApi.get(props.objectId, props.ruleId)
    rule.value = response.data
    form.value = {
      label: response.data.label,
      expression: response.data.expression,
      errorMessage: response.data.errorMessage,
      errorCode: response.data.errorCode,
      severity: response.data.severity,
      whenExpression: response.data.whenExpression ?? '',
      appliesTo: response.data.appliesTo,
      sortOrder: response.data.sortOrder,
      isActive: response.data.isActive,
      description: response.data.description,
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
function onSeverityChange(value: any) {
  form.value.severity = String(value) as 'error' | 'warning'
}

async function onSave() {
  submitting.value = true
  try {
    await validationRulesApi.update(props.objectId, props.ruleId, {
      label: form.value.label,
      expression: form.value.expression,
      errorMessage: form.value.errorMessage,
      errorCode: form.value.errorCode || undefined,
      severity: form.value.severity || undefined,
      whenExpression: form.value.whenExpression || undefined,
      appliesTo: form.value.appliesTo || undefined,
      sortOrder: form.value.sortOrder,
      isActive: form.value.isActive,
      description: form.value.description || undefined,
    })
    toast.success('Правило обновлено')
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDeleteRule() {
  try {
    await validationRulesApi.delete(props.objectId, props.ruleId)
    toast.success('Правило удалено')
    router.push({
      name: 'admin-validation-rules',
      params: { objectId: props.objectId },
    })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({
    name: 'admin-validation-rules',
    params: { objectId: props.objectId },
  })
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: 'Объект', to: `/admin/metadata/objects/${props.objectId}` },
  { label: 'Правила', to: `/admin/metadata/objects/${props.objectId}/rules` },
  { label: rule.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="loading && !rule" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="rule">
      <PageHeader :title="rule.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <Button
            variant="destructive"
            size="sm"
            data-testid="delete-rule-btn"
            @click="showDeleteDialog = true"
          >
            Удалить правило
          </Button>
        </template>
      </PageHeader>

      <form class="max-w-2xl space-y-6 mt-4" @submit.prevent="onSave">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="space-y-2">
              <Label>API Name</Label>
              <Input :model-value="rule.apiName" disabled />
            </div>

            <div class="space-y-2">
              <Label for="label">Название</Label>
              <Input id="label" v-model="form.label" required data-testid="field-label" />
            </div>

            <div class="space-y-2">
              <Label>CEL-выражение</Label>
              <ExpressionBuilder
                v-model="form.expression"
                context="validation_rule"
                :object-api-name="objectApiName || undefined"
                data-testid="field-expression"
              />
            </div>

            <div class="space-y-2">
              <Label for="error_message">Сообщение об ошибке</Label>
              <Input
                id="error_message"
                v-model="form.errorMessage"
                required
                data-testid="field-error-message"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="error_code">Код ошибки</Label>
                <Input id="error_code" v-model="form.errorCode" />
              </div>
              <div class="space-y-2">
                <Label>Серьёзность</Label>
                <Select :model-value="form.severity" @update:model-value="onSeverityChange">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="error">Ошибка</SelectItem>
                    <SelectItem value="warning">Предупреждение</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div class="space-y-2">
              <Label>Условие применения (CEL, необязательно)</Label>
              <ExpressionBuilder
                v-model="form.whenExpression"
                context="when_expression"
                :object-api-name="objectApiName || undefined"
                height="80px"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="applies_to">Применяется к</Label>
                <Input id="applies_to" v-model="form.appliesTo" placeholder="create,update" />
              </div>
              <div class="space-y-2">
                <Label for="sort_order">Порядок</Label>
                <Input id="sort_order" v-model.number="form.sortOrder" type="number" />
              </div>
            </div>

            <div class="flex items-center gap-2">
              <input
                id="is_active"
                v-model="form.isActive"
                type="checkbox"
                class="h-4 w-4"
                data-testid="field-is-active"
              />
              <Label for="is_active">Активно</Label>
            </div>

            <div class="space-y-2">
              <Label for="description">Описание</Label>
              <Textarea id="description" v-model="form.description" rows="2" />
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2">
          <Button type="submit" :disabled="submitting" data-testid="save-btn">
            Сохранить
          </Button>
          <Button variant="outline" type="button" data-testid="cancel-btn" @click="onCancel">
            Отмена
          </Button>
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить правило?"
        :description="`Правило «${rule.label}» будет удалено без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteRule"
      />
    </template>
  </div>
</template>
