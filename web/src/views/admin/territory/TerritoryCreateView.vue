<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { useTerritoryForm } from '@/composables/useTerritoryForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { storeToRefs } from 'pinia'

const router = useRouter()
const route = useRoute()
const store = useTerritoryAdminStore()
const toast = useToast()
const { territoriesLoading, territoriesError, territories, models } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useTerritoryForm()

onMounted(async () => {
  await store.fetchModels({ perPage: 1000 }).catch((err) => toast.errorFromApi(err))

  const queryModelId = route.query.modelId as string | undefined
  if (queryModelId) {
    state.modelId = queryModelId
    await store.fetchTerritories({ modelId: queryModelId, perPage: 1000 })
      .catch((err) => toast.errorFromApi(err))
  }
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onModelChange(value: any) {
  state.modelId = String(value)
  state.parentId = null
  store.fetchTerritories({ modelId: state.modelId, perPage: 1000 })
    .catch((err) => toast.errorFromApi(err))
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onParentChange(value: any) {
  state.parentId = value === '__none__' ? null : String(value)
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createTerritory(toCreateRequest())
    toast.success('Территория создана')
    router.push({ name: 'admin-territory-detail', params: { territoryId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Территории', to: '/admin/territory/territories' },
  { label: 'Новая территория' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать территорию" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="territoriesError" :message="territoriesError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="modelId">Модель</Label>
            <Select :model-value="state.modelId" @update:model-value="onModelChange">
              <SelectTrigger>
                <SelectValue placeholder="Выберите модель" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="model in models" :key="model.id" :value="model.id">
                  {{ model.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="errors.modelId" class="text-sm text-destructive">{{ errors.modelId }}</p>
          </div>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="north_america"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Название</Label>
            <Input id="label" v-model="state.label" placeholder="Северная Америка" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="parentId">Родительская территория</Label>
            <Select :model-value="state.parentId ?? '__none__'" @update:model-value="onParentChange">
              <SelectTrigger>
                <SelectValue placeholder="Без родителя" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">Без родителя</SelectItem>
                <SelectItem v-for="t in territories" :key="t.id" :value="t.id">
                  {{ t.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="description">Описание</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="territoriesLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
