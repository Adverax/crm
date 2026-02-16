<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useMetadataStore } from '@/stores/metadata'
import { useObjectForm } from '@/composables/useObjectForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ObjectFlagsSection from '@/components/admin/metadata/ObjectFlagsSection.vue'
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
import { computed } from 'vue'
import type { ObjectType, Visibility } from '@/types/metadata'

const router = useRouter()
const store = useMetadataStore()
const toast = useToast()
const { objectsLoading, objectsError } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useObjectForm()

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
function onObjectTypeChange(value: any) {
  state.objectType = String(value) as ObjectType
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onVisibilityChange(value: any) {
  state.visibility = String(value) as Visibility
}

async function onSubmit() {
  if (!validate()) return

  try {
    const created = await store.createObject(toCreateRequest())
    toast.success('Object created')
    router.push({ name: 'admin-object-detail', params: { objectId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Objects', to: '/admin/metadata/objects' },
  { label: 'New Object' },
]
</script>

<template>
  <div>
    <PageHeader title="Create Object" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="objectsError" :message="objectsError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">General Information</h2>

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
              <Label for="label">Label</Label>
              <Input id="label" v-model="state.label" placeholder="Invoice" />
              <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
            </div>
            <div class="space-y-2">
              <Label for="pluralLabel">Plural Label</Label>
              <Input id="pluralLabel" v-model="state.pluralLabel" placeholder="Invoices" />
              <p v-if="errors.pluralLabel" class="text-sm text-destructive">{{ errors.pluralLabel }}</p>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="objectType">Object Type</Label>
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
