<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMetadataStore } from '@/stores/metadata'
import { useObjectForm } from '@/composables/useObjectForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ObjectFlagsSection from '@/components/admin/metadata/ObjectFlagsSection.vue'
import ObjectTypeBadge from '@/components/admin/metadata/ObjectTypeBadge.vue'
import FieldsTable from '@/components/admin/metadata/FieldsTable.vue'
import FieldCreateDialog from '@/components/admin/metadata/FieldCreateDialog.vue'
import FieldEditDialog from '@/components/admin/metadata/FieldEditDialog.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { storeToRefs } from 'pinia'
import type { FieldDefinition, Visibility } from '@/types/metadata'

const props = defineProps<{
  objectId: string
}>()

const router = useRouter()
const store = useMetadataStore()
const toast = useToast()
const { currentObject, fields, objectsLoading, objectsError, fieldsLoading, fieldsError } = storeToRefs(store)

const { state, errors, validate, toUpdateRequest, initFrom } = useObjectForm()

const showDeleteDialog = ref(false)
const showFieldCreate = ref(false)
const editingField = ref<FieldDefinition | null>(null)

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
function onVisibilityChange(value: any) {
  state.visibility = String(value) as Visibility
}

async function loadData() {
  try {
    const obj = await store.fetchObject(props.objectId)
    initFrom(obj)
    await store.fetchFields(props.objectId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.objectId, loadData)

async function onSave() {
  if (!validate()) return
  try {
    await store.updateObject(props.objectId, toUpdateRequest())
    toast.success('Объект обновлён')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteObject() {
  try {
    await store.deleteObject(props.objectId)
    toast.success('Объект удалён')
    router.push({ name: 'admin-objects' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

async function onFieldCreated() {
  showFieldCreate.value = false
  toast.success('Поле создано')
  await store.fetchFields(props.objectId)
}

async function onFieldUpdated() {
  editingField.value = null
  toast.success('Поле обновлено')
  await store.fetchFields(props.objectId)
}

async function onFieldDelete(field: FieldDefinition) {
  try {
    await store.deleteField(props.objectId, field.id)
    toast.success('Поле удалено')
    await store.fetchFields(props.objectId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: currentObject.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="objectsLoading && !currentObject" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentObject">
      <PageHeader :title="currentObject.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <ObjectTypeBadge :type="currentObject.objectType" />
          <Button
            v-if="currentObject.isDeleteableObject && !currentObject.isPlatformManaged"
            variant="destructive"
            size="sm"
            @click="showDeleteDialog = true"
          >
            Удалить объект
          </Button>
        </template>
      </PageHeader>

      <ErrorAlert v-if="objectsError" :message="objectsError" class="mb-4" />
      <ErrorAlert v-if="fieldsError" :message="fieldsError" class="mb-4" />

      <Tabs default-value="info">
        <TabsList>
          <TabsTrigger value="info">Основное</TabsTrigger>
          <TabsTrigger value="fields">
            Поля ({{ fields.length }})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <form class="max-w-2xl space-y-6 mt-4" @submit.prevent="onSave">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Основная информация</h2>

                <div class="space-y-2">
                  <Label>API Name</Label>
                  <Input :model-value="state.apiName" disabled />
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <div class="space-y-2">
                    <Label for="label">Название</Label>
                    <Input id="label" v-model="state.label" />
                    <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
                  </div>
                  <div class="space-y-2">
                    <Label for="pluralLabel">Мн. число</Label>
                    <Input id="pluralLabel" v-model="state.pluralLabel" />
                    <p v-if="errors.pluralLabel" class="text-sm text-destructive">{{ errors.pluralLabel }}</p>
                  </div>
                </div>

                <div class="space-y-2">
                  <Label>Тип объекта</Label>
                  <Input :model-value="state.objectType" disabled />
                </div>

                <div class="space-y-2">
                  <Label for="visibility">Видимость (OWD)</Label>
                  <Select :model-value="state.visibility" @update:model-value="onVisibilityChange">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="private">Приватный</SelectItem>
                      <SelectItem value="public_read">Публичный (чтение)</SelectItem>
                      <SelectItem value="public_read_write">Публичный (чтение/запись)</SelectItem>
                      <SelectItem value="controlled_by_parent">Управляется родителем</SelectItem>
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
              <Button type="submit" :disabled="objectsLoading">
                Сохранить
              </Button>
              <Button variant="outline" type="button" @click="router.back()">
                Отмена
              </Button>
            </div>
          </form>
        </TabsContent>

        <TabsContent value="fields">
          <div class="mt-4">
            <FieldsTable
              :fields="fields"
              :loading="fieldsLoading"
              @create="showFieldCreate = true"
              @edit="editingField = $event"
              @delete="onFieldDelete"
            />
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить объект?"
        :description="`Объект «${currentObject.label}» (${currentObject.apiName}) и все его поля будут удалены без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteObject"
      />

      <FieldCreateDialog
        :open="showFieldCreate"
        :object-id="props.objectId"
        :objects="store.objects"
        @update:open="showFieldCreate = $event"
        @created="onFieldCreated"
      />

      <FieldEditDialog
        v-if="editingField"
        :open="!!editingField"
        :object-id="props.objectId"
        :field="editingField"
        :objects="store.objects"
        @update:open="editingField = null"
        @updated="onFieldUpdated"
      />
    </template>
  </div>
</template>
