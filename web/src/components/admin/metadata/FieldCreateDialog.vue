<script setup lang="ts">
import { ref } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import FieldConfigSection from './FieldConfigSection.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { useFieldForm } from '@/composables/useFieldForm'
import { useMetadataStore } from '@/stores/metadata'
import { useToast } from '@/composables/useToast'
import { FIELD_TYPE_LABELS, FIELD_SUBTYPE_LABELS } from '@/types/field-types'
import type { FieldType, FieldSubtype, ObjectDefinition } from '@/types/metadata'

const props = defineProps<{
  open: boolean
  objectId: string
  objects: ObjectDefinition[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  created: []
}>()

const store = useMetadataStore()
const toast = useToast()
const saving = ref(false)
const apiError = ref<string | null>(null)

const {
  state,
  errors,
  availableSubtypes,
  configFields,
  validate,
  onFieldTypeChange,
  toCreateRequest,
  reset,
} = useFieldForm()

const referencedObjectId = ref('')

function onClose() {
  reset()
  referencedObjectId.value = ''
  apiError.value = null
  emit('update:open', false)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function handleTypeChange(value: any) {
  onFieldTypeChange(String(value) as FieldType)
  referencedObjectId.value = ''
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function handleSubtypeChange(value: any) {
  state.fieldSubtype = String(value) as FieldSubtype
  state.config = {}
}

async function onSubmit() {
  if (!validate()) return
  saving.value = true
  apiError.value = null
  try {
    const req = toCreateRequest()
    if (referencedObjectId.value) {
      req.referencedObjectId = referencedObjectId.value
    }
    await store.createField(props.objectId, req)
    reset()
    referencedObjectId.value = ''
    emit('created')
  } catch (err) {
    if (err instanceof Error) {
      apiError.value = err.message
    }
    toast.errorFromApi(err)
  } finally {
    saving.value = false
  }
}

const isReferenceType = () => state.fieldType === 'reference'
</script>

<template>
  <Dialog :open="props.open" @update:open="onClose">
    <DialogContent class="max-w-lg max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create Field</DialogTitle>
        <DialogDescription>Add a new field for the object</DialogDescription>
      </DialogHeader>

      <ErrorAlert v-if="apiError" :message="apiError" class="mb-2" />

      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-2">
          <Label for="fieldApiName">API Name</Label>
          <Input id="fieldApiName" v-model="state.apiName" placeholder="invoice_number__c" />
          <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
        </div>

        <div class="space-y-2">
          <Label for="fieldLabel">Label</Label>
          <Input id="fieldLabel" v-model="state.label" placeholder="Invoice Number" />
          <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label>Field Type</Label>
            <Select :model-value="state.fieldType" @update:model-value="handleTypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="(label, key) in FIELD_TYPE_LABELS" :key="key" :value="key">
                  {{ label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div v-if="availableSubtypes.length > 0" class="space-y-2">
            <Label>Subtype</Label>
            <Select :model-value="state.fieldSubtype" @update:model-value="handleSubtypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="sub in availableSubtypes" :key="sub" :value="sub">
                  {{ FIELD_SUBTYPE_LABELS[sub] }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="fieldDescription">Description</Label>
          <Textarea id="fieldDescription" v-model="state.description" rows="2" />
        </div>

        <div class="space-y-2">
          <Label for="fieldHelpText">Help Text</Label>
          <Input id="fieldHelpText" v-model="state.helpText" />
        </div>

        <div class="flex items-center gap-6">
          <div class="flex items-center gap-2">
            <Switch id="fieldRequired" v-model:checked="state.isRequired" />
            <Label for="fieldRequired">Required</Label>
          </div>
          <div class="flex items-center gap-2">
            <Switch id="fieldUnique" v-model:checked="state.isUnique" />
            <Label for="fieldUnique">Unique</Label>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="fieldSortOrder">Sort Order</Label>
          <Input id="fieldSortOrder" type="number" v-model.number="state.sortOrder" />
        </div>

        <template v-if="configFields.length > 0 || isReferenceType()">
          <Separator />
          <h3 class="text-sm font-medium">Configuration</h3>
          <FieldConfigSection
            :config-fields="configFields"
            v-model="state.config"
            :show-referenced-object="isReferenceType()"
            :objects="props.objects"
            v-model:referenced-object-id="referencedObjectId"
          />
        </template>

        <DialogFooter>
          <Button variant="outline" type="button" @click="onClose">Cancel</Button>
          <Button type="submit" :disabled="saving">Create</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
