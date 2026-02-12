<script setup lang="ts">
import { onMounted, watch, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { useTerritoryModelForm } from '@/composables/useTerritoryModelForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { MODEL_STATUS_LABELS } from '@/types/territory'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  modelId: string
}>()

const router = useRouter()
const store = useTerritoryAdminStore()
const toast = useToast()
const { currentModel, modelsLoading, modelsError } = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = useTerritoryModelForm()

const showDeleteDialog = ref(false)
const showActivateDialog = ref(false)
const showArchiveDialog = ref(false)

async function loadData() {
  try {
    const model = await store.fetchModel(props.modelId)
    initFrom(model)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.modelId, loadData)

async function onSave() {
  if (!validate()) return
  try {
    await store.updateModel(props.modelId, toUpdateRequest())
    toast.success('Модель обновлена')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onActivate() {
  try {
    await store.activateModel(props.modelId)
    toast.success('Модель активирована')
    if (currentModel.value) initFrom(currentModel.value)
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showActivateDialog.value = false
  }
}

async function onArchive() {
  try {
    await store.archiveModel(props.modelId)
    toast.success('Модель архивирована')
    if (currentModel.value) initFrom(currentModel.value)
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showArchiveDialog.value = false
  }
}

async function onDelete() {
  try {
    await store.deleteModel(props.modelId)
    toast.success('Модель удалена')
    router.push({ name: 'admin-territory-models' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'active': return 'default'
    case 'archived': return 'secondary'
    default: return 'outline'
  }
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Модели территорий', to: '/admin/territory/models' },
  { label: currentModel.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="modelsLoading && !currentModel" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentModel">
      <PageHeader :title="currentModel.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <div class="flex gap-2">
            <Badge :variant="statusVariant(currentModel.status)" class="text-sm">
              {{ MODEL_STATUS_LABELS[currentModel.status] }}
            </Badge>
            <Button
              v-if="currentModel.status === 'planning'"
              variant="default"
              size="sm"
              @click="showActivateDialog = true"
            >
              Активировать
            </Button>
            <Button
              v-if="currentModel.status === 'active'"
              variant="secondary"
              size="sm"
              @click="showArchiveDialog = true"
            >
              Архивировать
            </Button>
            <Button
              variant="outline"
              size="sm"
              @click="router.push({ name: 'admin-territory-list', query: { modelId: props.modelId } })"
            >
              Территории
            </Button>
            <Button
              variant="destructive"
              size="sm"
              @click="showDeleteDialog = true"
            >
              Удалить
            </Button>
          </div>
        </template>
      </PageHeader>

      <ErrorAlert v-if="modelsError" :message="modelsError" class="mb-4" />

      <form class="max-w-2xl space-y-6" @submit.prevent="onSave">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <h2 class="text-lg font-semibold">Основная информация</h2>

            <div class="space-y-2">
              <Label>API Name</Label>
              <Input :model-value="state.apiName" disabled />
            </div>

            <div class="space-y-2">
              <Label for="label">Название</Label>
              <Input id="label" v-model="state.label" />
              <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
            </div>

            <div class="space-y-2">
              <Label for="description">Описание</Label>
              <Textarea id="description" v-model="state.description" rows="3" />
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2">
          <Button type="submit" :disabled="modelsLoading">
            Сохранить
          </Button>
          <Button variant="outline" type="button" @click="router.back()">
            Отмена
          </Button>
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить модель?"
        :description="`Модель «${currentModel.label}» (${currentModel.apiName}) будет удалена без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />

      <ConfirmDialog
        :open="showActivateDialog"
        title="Активировать модель?"
        :description="`Модель «${currentModel.label}» будет активирована. Территории станут участвовать в определении доступа к записям.`"
        @update:open="showActivateDialog = $event"
        @confirm="onActivate"
      />

      <ConfirmDialog
        :open="showArchiveDialog"
        title="Архивировать модель?"
        :description="`Модель «${currentModel.label}» будет архивирована. Территории перестанут влиять на доступ.`"
        @update:open="showArchiveDialog = $event"
        @confirm="onArchive"
      />
    </template>
  </div>
</template>
