<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { usePermissionEditorStore } from '@/stores/permissionEditor'
import { usePermissionSetForm } from '@/composables/usePermissionSetForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import PsTypeBadge from '@/components/admin/security/PsTypeBadge.vue'
import ObjectPermissionsEditor from '@/components/admin/security/ObjectPermissionsEditor.vue'
import FieldPermissionsEditor from '@/components/admin/security/FieldPermissionsEditor.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  permissionSetId: string
}>()

const router = useRouter()
const store = useSecurityAdminStore()
const permEditor = usePermissionEditorStore()
const toast = useToast()
const { currentPermissionSet, permissionSetsLoading, permissionSetsError } = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = usePermissionSetForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const ps = await store.fetchPermissionSet(props.permissionSetId)
    initFrom(ps)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

function loadPermissions() {
  permEditor.loadForPermissionSet(props.permissionSetId).catch((err) => toast.errorFromApi(err))
}

onMounted(() => {
  loadData()
  loadPermissions()
})

watch(() => props.permissionSetId, () => {
  permEditor.reset()
  loadData()
  loadPermissions()
})

onUnmounted(() => {
  permEditor.reset()
})

async function onSave() {
  if (!validate()) return
  try {
    await store.updatePermissionSet(props.permissionSetId, toUpdateRequest())
    toast.success('Набор разрешений обновлён')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeletePS() {
  try {
    await store.deletePermissionSet(props.permissionSetId)
    toast.success('Набор разрешений удалён')
    router.push({ name: 'admin-permission-sets' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Наборы разрешений', to: '/admin/security/permission-sets' },
  { label: currentPermissionSet.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="permissionSetsLoading && !currentPermissionSet" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentPermissionSet">
      <PageHeader :title="currentPermissionSet.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <PsTypeBadge :type="currentPermissionSet.psType" />
          <Button
            variant="destructive"
            size="sm"
            @click="showDeleteDialog = true"
          >
            Удалить набор
          </Button>
        </template>
      </PageHeader>

      <ErrorAlert v-if="permissionSetsError" :message="permissionSetsError" class="mb-4" />

      <Tabs default-value="info">
        <TabsList>
          <TabsTrigger value="info">Основное</TabsTrigger>
          <TabsTrigger value="ols">Права на объекты</TabsTrigger>
          <TabsTrigger value="fls">Права на поля</TabsTrigger>
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

                <div class="space-y-2">
                  <Label for="label">Название</Label>
                  <Input id="label" v-model="state.label" />
                  <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
                </div>

                <div class="space-y-2">
                  <Label>Тип</Label>
                  <Input :model-value="state.psType === 'grant' ? 'Grant' : 'Deny'" disabled />
                </div>

                <div class="space-y-2">
                  <Label for="description">Описание</Label>
                  <Textarea id="description" v-model="state.description" rows="3" />
                </div>
              </CardContent>
            </Card>

            <Separator />

            <div class="flex gap-2">
              <Button type="submit" :disabled="permissionSetsLoading">
                Сохранить
              </Button>
              <Button variant="outline" type="button" @click="router.back()">
                Отмена
              </Button>
            </div>
          </form>
        </TabsContent>

        <TabsContent value="ols">
          <div class="mt-4">
            <ObjectPermissionsEditor />
          </div>
        </TabsContent>

        <TabsContent value="fls">
          <div class="mt-4">
            <FieldPermissionsEditor />
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить набор разрешений?"
        :description="`Набор «${currentPermissionSet.label}» (${currentPermissionSet.apiName}) будет удалён без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeletePS"
      />
    </template>
  </div>
</template>
