<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { useTerritoryForm } from '@/composables/useTerritoryForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
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
    toast.success('Territory created')
    router.push({ name: 'admin-territory-detail', params: { territoryId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Territories', to: '/admin/territory/territories' },
  { label: 'New Territory' },
]
</script>

<template>
  <div>
    <PageHeader title="Create Territory" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="territoriesError" :message="territoriesError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">General Information</h2>

          <div class="space-y-2">
            <Label for="modelId">Model</Label>
            <Select :model-value="state.modelId" @update:model-value="onModelChange">
              <SelectTrigger>
                <SelectValue placeholder="Select model" />
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
            <Label for="label">Label</Label>
            <Input id="label" v-model="state.label" placeholder="North America" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="parentId">Parent Territory</Label>
            <Select :model-value="state.parentId ?? '__none__'" @update:model-value="onParentChange">
              <SelectTrigger>
                <SelectValue placeholder="No parent" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">No parent</SelectItem>
                <SelectItem v-for="t in territories" :key="t.id" :value="t.id">
                  {{ t.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="territoriesLoading">
          Create
        </Button>
        <IconButton
          :icon="X"
          tooltip="Cancel"
          variant="outline"
          @click="router.back()"
        />
      </div>
    </form>
  </div>
</template>
