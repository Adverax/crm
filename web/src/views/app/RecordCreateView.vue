<script setup lang="ts">
import { onMounted, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useRecordsStore } from '@/stores/records'
import { useFormFactor } from '@/composables/useFormFactor'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import FieldRenderer from '@/components/records/FieldRenderer.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import type { FieldDescribe } from '@/types/records'
import type { FormSection } from '@/types/object-views'

const props = defineProps<{ objectName: string }>()

const router = useRouter()
const store = useRecordsStore()
const formFactor = useFormFactor()
const toast = useToast()
const { currentDescribe, currentForm, loading, error } = storeToRefs(store)

const formData = reactive<Record<string, unknown>>({})

const hasSections = computed(() => {
  return (currentForm.value?.sections?.length ?? 0) > 0
})

function sectionFields(section: FormSection): FieldDescribe[] {
  if (!section.fields) return []
  return store.resolveFields(section.fields).filter(
    (f) => !f.isSystemField && !f.isReadOnly,
  )
}

onMounted(async () => {
  try {
    await store.fetchDescribe(props.objectName, { formFactor: formFactor.value, formMode: 'edit' })
    if (currentDescribe.value) {
      for (const field of store.editableFields) {
        if (field.config.defaultValue != null) {
          formData[field.apiName] = field.config.defaultValue
        }
      }
    }
  } catch (err) {
    toast.errorFromApi(err)
  }
})

async function onSubmit() {
  try {
    const data: Record<string, unknown> = {}
    for (const field of store.editableFields) {
      const val = formData[field.apiName]
      if (val !== undefined && val !== '' && val !== null) {
        data[field.apiName] = val
      }
    }
    const id = await store.createRecord(props.objectName, data)
    toast.success('Record created')
    router.push({ name: 'record-detail', params: { objectName: props.objectName, recordId: id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = computed(() => [
  { label: 'CRM', to: '/app' },
  { label: currentDescribe.value?.pluralLabel ?? props.objectName, to: `/app/${props.objectName}` },
  { label: 'New Record' },
])
</script>

<template>
  <div>
    <PageHeader
      :title="`Create ${currentDescribe?.label ?? objectName}`"
      :breadcrumbs="breadcrumbs"
    />

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <!-- Section-based layout when form has sections -->
      <template v-if="hasSections">
        <Card
          v-for="section in currentForm!.sections"
          :key="section.key"
          class="mb-4"
        >
          <CardHeader class="py-3">
            <CardTitle class="text-base">{{ section.label }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              :class="section.columns === 2 ? 'grid grid-cols-2 gap-4' : 'space-y-4'"
            >
              <FieldRenderer
                v-for="field in sectionFields(section)"
                :key="field.apiName"
                :field="field"
                :model-value="formData[field.apiName]"
                @update:model-value="formData[field.apiName] = $event"
              />
            </div>
          </CardContent>
        </Card>
      </template>

      <!-- Fallback: single card with all editable fields -->
      <Card v-else>
        <CardContent class="pt-6 space-y-4">
          <FieldRenderer
            v-for="field in store.editableFields"
            :key="field.apiName"
            :field="field"
            :model-value="formData[field.apiName]"
            @update:model-value="formData[field.apiName] = $event"
          />
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="loading">
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
