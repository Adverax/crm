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
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
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
import { RouterLink } from 'vue-router'
import type { FieldDefinition, Visibility } from '@/types/metadata'

const props = defineProps<{
  objectId: string
}>()

const router = useRouter()
const store = useMetadataStore()
const toast = useToast()
const { currentObject, fields, objectsLoading, objectsError, fieldsLoading, fieldsError } = storeToRefs(store)

const { state, errors, validate, toUpdateRequest, initFrom } = useObjectForm()

const activeTab = ref('info')
const showDeleteDialog = ref(false)
const showFieldCreate = ref(false)
const editingField = ref<FieldDefinition | null>(null)

const flagGroups = [
  {
    title: 'Record Permissions',
    items: [
      { key: 'isCreateable', label: 'Create records' },
      { key: 'isUpdateable', label: 'Update records' },
      { key: 'isDeleteable', label: 'Delete records' },
      { key: 'isQueryable', label: 'Queries (SOQL)' },
      { key: 'isSearchable', label: 'Full-text search' },
    ],
  },
  {
    title: 'Object Settings',
    items: [
      { key: 'isVisibleInSetup', label: 'Visible in setup' },
      { key: 'isCustomFieldsAllowed', label: 'Custom fields allowed' },
      { key: 'isDeleteableObject', label: 'Object can be deleted' },
    ],
  },
  {
    title: 'Capabilities',
    items: [
      { key: 'hasActivities', label: 'Activities' },
      { key: 'hasNotes', label: 'Notes' },
      { key: 'hasHistoryTracking', label: 'History tracking' },
      { key: 'hasSharingRules', label: 'Sharing rules' },
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
    await Promise.all([
      store.fetchFields(props.objectId),
      store.fetchObjects(),
    ])
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
    toast.success('Object updated')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteObject() {
  try {
    await store.deleteObject(props.objectId)
    toast.success('Object deleted')
    router.push({ name: 'admin-objects' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

async function onFieldCreated() {
  showFieldCreate.value = false
  toast.success('Field created')
  await store.fetchFields(props.objectId)
}

async function onFieldUpdated() {
  editingField.value = null
  toast.success('Field updated')
  await store.fetchFields(props.objectId)
}

async function onFieldDelete(field: FieldDefinition) {
  try {
    await store.deleteField(props.objectId, field.id)
    toast.success('Field deleted')
    await store.fetchFields(props.objectId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const tabLabels: Record<string, string> = {
  info: 'General',
  fields: 'Fields',
}

const breadcrumbs = computed(() => {
  const crumbs: { label: string; to?: string }[] = [
    { label: 'Admin', to: '/admin' },
    { label: 'Objects', to: '/admin/metadata/objects' },
  ]
  if (activeTab.value !== 'info') {
    crumbs.push({ label: currentObject.value?.label ?? '...', to: `/admin/metadata/objects/${props.objectId}` })
    crumbs.push({ label: tabLabels[activeTab.value] ?? activeTab.value })
  } else {
    crumbs.push({ label: currentObject.value?.label ?? '...' })
  }
  return crumbs
})
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
          <IconButton
            v-if="currentObject.isDeleteableObject && !currentObject.isPlatformManaged"
            :icon="Trash2"
            tooltip="Delete object"
            variant="destructive"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <ErrorAlert v-if="objectsError" :message="objectsError" class="mb-4" />
      <ErrorAlert v-if="fieldsError" :message="fieldsError" class="mb-4" />

      <Tabs v-model="activeTab">
        <TabsList>
          <TabsTrigger value="info">General</TabsTrigger>
          <TabsTrigger value="fields">
            Fields ({{ fields.length }})
          </TabsTrigger>
          <TabsTrigger value="rules" as-child>
            <RouterLink
              :to="{ name: 'admin-validation-rules', params: { objectId: props.objectId } }"
              class="inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium"
            >
              Validation Rules
            </RouterLink>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <form class="max-w-2xl space-y-6 mt-4" @submit.prevent="onSave">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">General Information</h2>

                <div class="space-y-2">
                  <Label>API Name</Label>
                  <Input :model-value="state.apiName" disabled />
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <div class="space-y-2">
                    <Label for="label">Label</Label>
                    <Input id="label" v-model="state.label" />
                    <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
                  </div>
                  <div class="space-y-2">
                    <Label for="pluralLabel">Plural Label</Label>
                    <Input id="pluralLabel" v-model="state.pluralLabel" />
                    <p v-if="errors.pluralLabel" class="text-sm text-destructive">{{ errors.pluralLabel }}</p>
                  </div>
                </div>

                <div class="space-y-2">
                  <Label>Object Type</Label>
                  <Input :model-value="state.objectType" disabled />
                </div>

                <div class="space-y-2">
                  <Label for="visibility">Visibility (OWD)</Label>
                  <Select :model-value="state.visibility" @update:model-value="onVisibilityChange">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="private">Private</SelectItem>
                      <SelectItem value="public_read">Public Read Only</SelectItem>
                      <SelectItem value="public_read_write">Public Read/Write</SelectItem>
                      <SelectItem value="controlled_by_parent">Controlled by Parent</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div class="space-y-2">
                  <Label for="description">Description</Label>
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

            <div class="flex gap-2 items-center">
              <Button type="submit" :disabled="objectsLoading">
                Save
              </Button>
              <IconButton
                :icon="X"
                tooltip="Cancel"
                variant="outline"
                @click="router.back()"
              />
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
        title="Delete object?"
        :description="`Object '${currentObject.label}' (${currentObject.apiName}) and all its fields will be permanently deleted.`"
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
