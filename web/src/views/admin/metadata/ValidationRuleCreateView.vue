<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { validationRulesApi } from '@/api/validationRules'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
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

const props = defineProps<{
  objectId: string
}>()

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const form = ref({
  apiName: '',
  label: '',
  expression: '',
  errorMessage: '',
  errorCode: 'validation_failed',
  severity: 'error' as 'error' | 'warning',
  whenExpression: '',
  appliesTo: 'create,update',
  description: '',
})

async function onSubmit() {
  submitting.value = true
  try {
    await validationRulesApi.create(props.objectId, {
      apiName: form.value.apiName,
      label: form.value.label,
      expression: form.value.expression,
      errorMessage: form.value.errorMessage,
      errorCode: form.value.errorCode || 'validation_failed',
      severity: form.value.severity || 'error',
      whenExpression: form.value.whenExpression || undefined,
      appliesTo: form.value.appliesTo || 'create,update',
      sortOrder: 0,
      isActive: true,
      description: form.value.description || undefined,
    })
    toast.success('Правило создано')
    router.push({
      name: 'admin-validation-rules',
      params: { objectId: props.objectId },
    })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({
    name: 'admin-validation-rules',
    params: { objectId: props.objectId },
  })
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onSeverityChange(value: any) {
  form.value.severity = String(value) as 'error' | 'warning'
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: 'Объект', to: `/admin/metadata/objects/${props.objectId}` },
  { label: 'Правила', to: `/admin/metadata/objects/${props.objectId}/rules` },
  { label: 'Создание' },
])
</script>

<template>
  <div>
    <PageHeader title="Создание правила валидации" :breadcrumbs="breadcrumbs" />

    <form class="max-w-2xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="api_name">API Name</Label>
              <Input id="api_name" v-model="form.apiName" required data-testid="field-api-name" />
            </div>
            <div class="space-y-2">
              <Label for="label">Название</Label>
              <Input id="label" v-model="form.label" required data-testid="field-label" />
            </div>
          </div>

          <div class="space-y-2">
            <Label for="expression">CEL-выражение</Label>
            <Textarea
              id="expression"
              v-model="form.expression"
              rows="3"
              required
              placeholder='size(record.Name) > 0'
              data-testid="field-expression"
            />
            <p class="text-xs text-muted-foreground">
              Выражение должно возвращать true, если данные корректны.
            </p>
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
            <Label for="when_expression">Условие применения (CEL, необязательно)</Label>
            <Textarea
              id="when_expression"
              v-model="form.whenExpression"
              rows="2"
              placeholder='record.Type == "Premium"'
            />
          </div>

          <div class="space-y-2">
            <Label for="applies_to">Применяется к</Label>
            <Input id="applies_to" v-model="form.appliesTo" placeholder="create,update" />
          </div>

          <div class="space-y-2">
            <Label for="description">Описание</Label>
            <Textarea id="description" v-model="form.description" rows="2" />
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="onCancel" data-testid="cancel-btn">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
