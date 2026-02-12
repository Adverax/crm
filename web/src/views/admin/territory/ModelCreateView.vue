<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { useTerritoryModelForm } from '@/composables/useTerritoryModelForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useTerritoryAdminStore()
const toast = useToast()
const { modelsLoading, modelsError } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useTerritoryModelForm()

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createModel(toCreateRequest())
    toast.success('Модель создана')
    router.push({ name: 'admin-territory-model-detail', params: { modelId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Модели территорий', to: '/admin/territory/models' },
  { label: 'Новая модель' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать модель территорий" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="modelsError" :message="modelsError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="q1_2026"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Название</Label>
            <Input id="label" v-model="state.label" placeholder="Q1 2026" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="description">Описание</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="modelsLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
