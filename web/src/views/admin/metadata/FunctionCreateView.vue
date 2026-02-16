<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { functionsApi } from '@/api/functions'
import { useToast } from '@/composables/useToast'
import { useFunctionsStore } from '@/stores/functions'
import PageHeader from '@/components/admin/PageHeader.vue'
import ExpressionBuilder from '@/components/admin/expression-builder/ExpressionBuilder.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, Trash2, X } from 'lucide-vue-next'
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
import type { FunctionParam } from '@/types/functions'

const router = useRouter()
const toast = useToast()
const functionsStore = useFunctionsStore()
const submitting = ref(false)

const form = ref({
  name: '',
  description: '',
  returnType: 'any' as string,
  body: '',
  params: [] as FunctionParam[],
})

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

async function onSubmit() {
  submitting.value = true
  try {
    await functionsApi.create({
      name: form.value.name,
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
    toast.success('Function created')
    await functionsStore.invalidate()
    router.push({ name: 'admin-functions' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-functions' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Functions', to: '/admin/metadata/functions' },
  { label: 'Create' },
])

const functionParams = computed(() =>
  form.value.params.filter((p) => p.name),
)
</script>

<template>
  <div>
    <PageHeader title="Create Function" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="name">Function Name</Label>
              <Input
                id="name"
                v-model="form.name"
                required
                placeholder="my_function"
                class="font-mono"
                data-testid="field-name"
              />
              <p class="text-xs text-muted-foreground">
                Called as fn.{{ form.name || 'name' }}()
              </p>
            </div>
            <div class="space-y-2">
              <Label>Return Type</Label>
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
            <Label for="description">Description</Label>
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
            <Label class="text-base">Parameters</Label>
            <IconButton
              :icon="Plus"
              tooltip="Add parameter"
              variant="outline"
              data-testid="add-param-btn"
              @click="addParam"
            />
          </div>

          <div
            v-for="(param, idx) in form.params"
            :key="idx"
            class="grid grid-cols-[1fr_120px_1fr_auto] gap-2 items-end"
            data-testid="param-row"
          >
            <div class="space-y-1">
              <Label class="text-xs">Name</Label>
              <Input
                v-model="param.name"
                required
                placeholder="x"
                class="font-mono"
                :data-testid="`param-name-${idx}`"
              />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">Type</Label>
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
              <Label class="text-xs">Description</Label>
              <Input
                v-model="param.description"
                placeholder="Description"
                :data-testid="`param-desc-${idx}`"
              />
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              :data-testid="`remove-param-${idx}`"
              @click="removeParam(idx)"
            />
          </div>

          <div v-if="form.params.length === 0" class="text-sm text-muted-foreground">
            No parameters. Function will be called as fn.{{ form.name || 'name' }}()
          </div>
        </CardContent>
      </Card>

      <!-- Body -->
      <Card>
        <CardContent class="pt-6 space-y-4">
          <Label>Function Body (CEL expression)</Label>
          <ExpressionBuilder
            v-model="form.body"
            context="function_body"
            :function-params="functionParams"
            height="160px"
            placeholder="x * 2 + 1"
            :show-field-picker="form.params.length > 0"
            data-testid="field-body"
          />
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton
          :icon="X"
          tooltip="Cancel"
          variant="outline"
          data-testid="cancel-btn"
          @click="onCancel"
        />
      </div>
    </form>
  </div>
</template>
