<script setup lang="ts">
import { onMounted, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useRecordsStore } from '@/stores/records'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import FieldRenderer from '@/components/records/FieldRenderer.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

const props = defineProps<{ objectName: string }>()

const router = useRouter()
const store = useRecordsStore()
const toast = useToast()
const { currentDescribe, loading, error } = storeToRefs(store)

const formData = reactive<Record<string, unknown>>({})

onMounted(async () => {
  try {
    await store.fetchDescribe(props.objectName)
    // Pre-fill defaults from metadata
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
