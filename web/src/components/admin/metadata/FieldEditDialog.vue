<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
import { Separator } from '@/components/ui/separator'
import FieldConfigSection from './FieldConfigSection.vue'
import FieldTypeBadge from './FieldTypeBadge.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { useFieldForm } from '@/composables/useFieldForm'
import { useMetadataStore } from '@/stores/metadata'
import { useToast } from '@/composables/useToast'
import type { FieldDefinition, ObjectDefinition } from '@/types/metadata'

const props = defineProps<{
  open: boolean
  objectId: string
  field: FieldDefinition
  objects: ObjectDefinition[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: []
}>()

const store = useMetadataStore()
const toast = useToast()
const saving = ref(false)
const apiError = ref<string | null>(null)

const {
  state,
  errors,
  configFields,
  validate,
  toUpdateRequest,
  initFrom,
} = useFieldForm(props.field)

const referencedObjectId = ref(props.field.referencedObjectId ?? '')

onMounted(() => {
  initFrom(props.field)
  referencedObjectId.value = props.field.referencedObjectId ?? ''
})

function onClose() {
  apiError.value = null
  emit('update:open', false)
}

async function onSubmit() {
  if (!validate()) return
  saving.value = true
  apiError.value = null
  try {
    await store.updateField(props.objectId, props.field.id, toUpdateRequest())
    emit('updated')
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
        <DialogTitle>Edit Field</DialogTitle>
        <DialogDescription>
          {{ props.field.apiName }}
        </DialogDescription>
      </DialogHeader>

      <ErrorAlert v-if="apiError" :message="apiError" class="mb-2" />

      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-2">
          <Label>API Name</Label>
          <Input :model-value="state.apiName" disabled />
        </div>

        <div class="space-y-2">
          <Label>Type</Label>
          <div>
            <FieldTypeBadge :field-type="state.fieldType" :field-subtype="state.fieldSubtype || undefined" />
          </div>
        </div>

        <div class="space-y-2">
          <Label for="editFieldLabel">Label</Label>
          <Input id="editFieldLabel" v-model="state.label" />
          <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
        </div>

        <div class="space-y-2">
          <Label for="editFieldDescription">Description</Label>
          <Textarea id="editFieldDescription" v-model="state.description" rows="2" />
        </div>

        <div class="space-y-2">
          <Label for="editFieldHelpText">Help Text</Label>
          <Input id="editFieldHelpText" v-model="state.helpText" />
        </div>

        <div class="flex items-center gap-6">
          <div class="flex items-center gap-2">
            <Switch id="editFieldRequired" v-model:checked="state.isRequired" />
            <Label for="editFieldRequired">Required</Label>
          </div>
          <div class="flex items-center gap-2">
            <Switch id="editFieldUnique" v-model:checked="state.isUnique" />
            <Label for="editFieldUnique">Unique</Label>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="editFieldSortOrder">Sort Order</Label>
          <Input id="editFieldSortOrder" type="number" v-model.number="state.sortOrder" />
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
          <Button type="submit" :disabled="saving">Save</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
