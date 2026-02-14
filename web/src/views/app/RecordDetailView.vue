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
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

const props = defineProps<{
  objectName: string
  recordId: string
}>()

const router = useRouter()
const store = useRecordsStore()
const toast = useToast()
const { currentDescribe, currentRecord, loading, error } = storeToRefs(store)

const formData = reactive<Record<string, unknown>>({})
const showDeleteDialog = ref(false)

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
    toast.success('Запись обновлена')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteConfirmed() {
  try {
    await store.deleteRecord(props.objectName, props.recordId)
    toast.success('Запись удалена')
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
        <Button
          v-if="currentDescribe?.isDeleteable"
          variant="destructive"
          @click="showDeleteDialog = true"
        >
          Удалить
        </Button>
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSave">
      <Card>
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

      <div class="flex gap-2">
        <Button
          v-if="currentDescribe?.isUpdateable"
          type="submit"
          :disabled="loading"
        >
          Сохранить
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Удалить запись?"
      description="Запись будет удалена без возможности восстановления."
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
