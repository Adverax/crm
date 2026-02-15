<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { functionsApi } from '@/api/functions'
import { useToast } from '@/composables/useToast'
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
import type { Function, FunctionParam } from '@/types/functions'

const props = defineProps<{
  functionId: string
}>()

const router = useRouter()
const toast = useToast()

const fn = ref<Function | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)

const form = ref({
  description: '',
  returnType: 'any' as string,
  body: '',
  params: [] as FunctionParam[],
})

async function loadFunction() {
  loading.value = true
  try {
    const response = await functionsApi.get(props.functionId)
    fn.value = response.data
    form.value = {
      description: response.data.description ?? '',
      returnType: response.data.returnType ?? 'any',
      body: response.data.body ?? '',
      params: (response.data.params ?? []).map((p) => ({
        name: p.name,
        type: p.type,
        description: p.description ?? '',
      })),
    }
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadFunction)
watch(() => props.functionId, loadFunction)

function addParam() {
  form.value.params.push({ name: '', type: 'any', description: '' })
}

function removeParam(index: number) {
  form.value.params.splice(index, 1)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onReturnTypeChange(value: any) {
  form.value.returnType = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onParamTypeChange(index: number, value: any) {
  const param = form.value.params[index]
  if (param) {
    param.type = String(value) as FunctionParam['type']
  }
}

async function onSave() {
  submitting.value = true
  try {
    await functionsApi.update(props.functionId, {
      description: form.value.description || undefined,
      returnType: form.value.returnType as 'string' | 'number' | 'boolean' | 'list' | 'map' | 'any' | undefined,
      body: form.value.body,
      params: form.value.params.length > 0
        ? form.value.params.map((p) => ({
            name: p.name,
            type: p.type,
            description: p.description || undefined,
          }))
        : undefined,
    })
    toast.success('Функция обновлена')
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await functionsApi.delete(props.functionId)
    toast.success('Функция удалена')
    router.push({ name: 'admin-functions' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-functions' })
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Функции', to: '/admin/metadata/functions' },
  { label: fn.value ? `fn.${fn.value.name}` : '...' },
])

const functionParams = computed(() =>
  form.value.params.filter((p) => p.name),
)
</script>

<template>
  <div>
    <div v-if="loading && !fn" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="fn">
      <PageHeader :title="`fn.${fn.name}`" :breadcrumbs="breadcrumbs">
        <template #actions>
          <Button
            variant="destructive"
            size="sm"
            data-testid="delete-function-btn"
            @click="showDeleteDialog = true"
          >
            Удалить функцию
          </Button>
        </template>
      </PageHeader>

      <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSave">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label>Имя функции</Label>
                <Input :model-value="fn.name" disabled class="font-mono" />
              </div>
              <div class="space-y-2">
                <Label>Тип возврата</Label>
                <Select :model-value="form.returnType" @update:model-value="onReturnTypeChange">
                  <SelectTrigger data-testid="field-return-type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="any">any</SelectItem>
                    <SelectItem value="string">string</SelectItem>
                    <SelectItem value="number">number</SelectItem>
                    <SelectItem value="boolean">boolean</SelectItem>
                    <SelectItem value="list">list</SelectItem>
                    <SelectItem value="map">map</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="description">Описание</Label>
              <Textarea
                id="description"
                v-model="form.description"
                rows="2"
                data-testid="field-description"
              />
            </div>
          </CardContent>
        </Card>

        <!-- Parameters -->
        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="flex items-center justify-between">
              <Label class="text-base">Параметры</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                data-testid="add-param-btn"
                @click="addParam"
              >
                Добавить параметр
              </Button>
            </div>

            <div
              v-for="(param, idx) in form.params"
              :key="idx"
              class="grid grid-cols-[1fr_120px_1fr_auto] gap-2 items-end"
              data-testid="param-row"
            >
              <div class="space-y-1">
                <Label class="text-xs">Имя</Label>
                <Input
                  v-model="param.name"
                  required
                  placeholder="x"
                  class="font-mono"
                  :data-testid="`param-name-${idx}`"
                />
              </div>
              <div class="space-y-1">
                <Label class="text-xs">Тип</Label>
                <Select :model-value="param.type" @update:model-value="(v) => onParamTypeChange(idx, v)">
                  <SelectTrigger :data-testid="`param-type-${idx}`">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="any">any</SelectItem>
                    <SelectItem value="string">string</SelectItem>
                    <SelectItem value="number">number</SelectItem>
                    <SelectItem value="boolean">boolean</SelectItem>
                    <SelectItem value="list">list</SelectItem>
                    <SelectItem value="map">map</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="space-y-1">
                <Label class="text-xs">Описание</Label>
                <Input
                  v-model="param.description"
                  placeholder="Описание"
                  :data-testid="`param-desc-${idx}`"
                />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="text-destructive"
                :data-testid="`remove-param-${idx}`"
                @click="removeParam(idx)"
              >
                Удалить
              </Button>
            </div>

            <div v-if="form.params.length === 0" class="text-sm text-muted-foreground">
              Нет параметров. Функция вызывается как fn.{{ fn.name }}()
            </div>
          </CardContent>
        </Card>

        <!-- Body -->
        <Card>
          <CardContent class="pt-6 space-y-4">
            <Label>Тело функции (CEL-выражение)</Label>
            <ExpressionBuilder
              v-model="form.body"
              context="function_body"
              :function-params="functionParams"
              height="160px"
              :show-field-picker="form.params.length > 0"
            />
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
        title="Удалить функцию?"
        :description="`Функция «fn.${fn.name}» будет удалена без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
