<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useMetadataStore } from '@/stores/metadata'
import { useObjectForm } from '@/composables/useObjectForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ObjectFlagsSection from '@/components/admin/metadata/ObjectFlagsSection.vue'
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
import { computed } from 'vue'
import type { ObjectType } from '@/types/metadata'

const router = useRouter()
const store = useMetadataStore()
const toast = useToast()
const { isLoading, error } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useObjectForm()

const flagGroups = [
  {
    title: 'Разрешения на записи',
    items: [
      { key: 'isCreateable', label: 'Создание записей' },
      { key: 'isUpdateable', label: 'Обновление записей' },
      { key: 'isDeleteable', label: 'Удаление записей' },
      { key: 'isQueryable', label: 'Запросы (SOQL)' },
      { key: 'isSearchable', label: 'Полнотекстовый поиск' },
    ],
  },
  {
    title: 'Настройки объекта',
    items: [
      { key: 'isVisibleInSetup', label: 'Виден в настройках' },
      { key: 'isCustomFieldsAllowed', label: 'Разрешены custom-поля' },
      { key: 'isDeleteableObject', label: 'Можно удалить объект' },
    ],
  },
  {
    title: 'Возможности',
    items: [
      { key: 'hasActivities', label: 'Активности' },
      { key: 'hasNotes', label: 'Заметки' },
      { key: 'hasHistoryTracking', label: 'История изменений' },
      { key: 'hasSharingRules', label: 'Правила общего доступа' },
    ],
  },
]

const flagsModel = computed({
  get: () => ({
    isCreateable: state.isCreateable,
    isUpdateable: state.isUpdateable,
    isDeleteable: state.isDeleteable,
    isQueryable: state.isQueryable,
    isSearchable: state.isSearchable,
    isVisibleInSetup: state.isVisibleInSetup,
    isCustomFieldsAllowed: state.isCustomFieldsAllowed,
    isDeleteableObject: state.isDeleteableObject,
    hasActivities: state.hasActivities,
    hasNotes: state.hasNotes,
    hasHistoryTracking: state.hasHistoryTracking,
    hasSharingRules: state.hasSharingRules,
  }),
  set: (val: Record<string, boolean>) => {
    Object.assign(state, val)
  },
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onObjectTypeChange(value: any) {
  state.objectType = String(value) as ObjectType
}

async function onSubmit() {
  if (!validate()) return

  try {
    const created = await store.createObject(toCreateRequest())
    toast.success('Объект создан')
    router.push({ name: 'admin-object-detail', params: { objectId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: 'Новый объект' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать объект" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="Invoice__c"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="label">Название</Label>
              <Input id="label" v-model="state.label" placeholder="Счёт" />
              <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
            </div>
            <div class="space-y-2">
              <Label for="pluralLabel">Мн. число</Label>
              <Input id="pluralLabel" v-model="state.pluralLabel" placeholder="Счета" />
              <p v-if="errors.pluralLabel" class="text-sm text-destructive">{{ errors.pluralLabel }}</p>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="objectType">Тип объекта</Label>
            <Select :model-value="state.objectType" @update:model-value="onObjectTypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="standard">Standard</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="description">Описание</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6">
          <ObjectFlagsSection
            :groups="flagGroups"
            v-model="flagsModel"
          />
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="isLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
