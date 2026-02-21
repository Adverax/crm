<script setup lang="ts">
import { onMounted, reactive, watch, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useRecordsStore } from '@/stores/records'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import FieldRenderer from '@/components/records/FieldRenderer.vue'
import FieldDisplay from '@/components/records/FieldDisplay.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X, ChevronDown } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import type { FieldDescribe } from '@/types/records'
import type { FormSection } from '@/types/object-views'

const props = defineProps<{
  objectName: string
  recordId: string
}>()

const router = useRouter()
const store = useRecordsStore()
const toast = useToast()
const { currentDescribe, currentRecord, currentForm, loading, error } = storeToRefs(store)

const formData = reactive<Record<string, unknown>>({})
const showDeleteDialog = ref(false)
const collapsedSections = ref<Set<string>>(new Set())

const hasSections = computed(() => {
  return (currentForm.value?.sections?.length ?? 0) > 0
})

function sectionFields(section: FormSection): FieldDescribe[] {
  if (!section.fields) return []
  return store.resolveFields(section.fields).filter(
    (f) => !f.isSystemField && !f.isReadOnly,
  )
}

function isSectionOpen(key: string, collapsed?: boolean): boolean {
  if (collapsedSections.value.has(key)) return false
  return !collapsed
}

function toggleSection(key: string, collapsed?: boolean) {
  if (isSectionOpen(key, collapsed)) {
    collapsedSections.value.add(key)
  } else {
    collapsedSections.value.delete(key)
  }
}

onMounted(loadData)

async function loadData() {
  try {
    await Promise.all([
      store.fetchDescribe(props.objectName),
      store.fetchRecord(props.objectName, props.recordId),
    ])
    syncFormData()
  } catch (err) {
    toast.errorFromApi(err)
  }
}

function syncFormData() {
  if (!currentRecord.value) return
  for (const field of store.editableFields) {
    formData[field.apiName] = currentRecord.value[field.apiName] ?? null
  }
}

watch(currentRecord, syncFormData)

async function onSave() {
  try {
    const data: Record<string, unknown> = {}
    for (const field of store.editableFields) {
      data[field.apiName] = formData[field.apiName]
    }
    await store.updateRecord(props.objectName, props.recordId, data)
    toast.success('Record updated')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteConfirmed() {
  try {
    await store.deleteRecord(props.objectName, props.recordId)
    toast.success('Record deleted')
    router.push({ name: 'record-list', params: { objectName: props.objectName } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function getRecordTitle(): string {
  if (!currentRecord.value) return '...'
  const name = currentRecord.value['Name'] ?? currentRecord.value['Id'] ?? props.recordId
  return String(name)
}

const breadcrumbs = computed(() => [
  { label: 'CRM', to: '/app' },
  { label: currentDescribe.value?.pluralLabel ?? props.objectName, to: `/app/${props.objectName}` },
  { label: getRecordTitle() },
])
</script>

<template>
  <div>
    <PageHeader :title="getRecordTitle()" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          v-if="currentDescribe?.isDeleteable"
          :icon="Trash2"
          tooltip="Delete"
          variant="destructive"
          @click="showDeleteDialog = true"
        />
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSave">
      <!-- Highlight fields -->
      <Card v-if="store.formHighlightFields.length && currentRecord" class="mb-4">
        <CardContent class="pt-4">
          <div class="grid grid-cols-3 gap-4">
            <div v-for="field in store.formHighlightFields" :key="field.apiName">
              <span class="text-xs text-muted-foreground">{{ field.label }}</span>
              <div class="font-medium">
                <FieldDisplay :field="field" :value="currentRecord[field.apiName]" />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Section-based layout when form has sections -->
      <template v-if="hasSections">
        <Card
          v-for="section in currentForm!.sections"
          :key="section.key"
          class="mb-4"
        >
          <CardHeader
            class="cursor-pointer select-none py-3"
            @click="toggleSection(section.key!, section.collapsed)"
          >
            <div class="flex items-center justify-between">
              <CardTitle class="text-base">{{ section.label }}</CardTitle>
              <ChevronDown
                class="h-4 w-4 transition-transform"
                :class="{ '-rotate-180': !isSectionOpen(section.key!, section.collapsed) }"
              />
            </div>
          </CardHeader>
          <CardContent
            v-show="isSectionOpen(section.key!, section.collapsed)"
          >
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
        <Button
          v-if="currentDescribe?.isUpdateable"
          type="submit"
          :disabled="loading"
        >
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

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete record?"
      description="The record will be permanently deleted."
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
