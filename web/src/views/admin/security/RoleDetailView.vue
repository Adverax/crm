<script setup lang="ts">
import { onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useRoleForm } from '@/composables/useRoleForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
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
import { Skeleton } from '@/components/ui/skeleton'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'

const props = defineProps<{
  roleId: string
}>()

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { currentRole, roles, rolesLoading, rolesError } = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = useRoleForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const role = await store.fetchRole(props.roleId)
    initFrom(role)
    await store.fetchRoles({ perPage: 1000 })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.roleId, loadData)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onParentChange(value: any) {
  state.parentId = value === '__none__' ? null : String(value)
}

async function onSave() {
  if (!validate()) return
  try {
    await store.updateRole(props.roleId, toUpdateRequest())
    toast.success('Роль обновлена')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteRole() {
  try {
    await store.deleteRole(props.roleId)
    toast.success('Роль удалена')
    router.push({ name: 'admin-roles' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const availableParents = computed(() =>
  roles.value.filter((r) => r.id !== props.roleId),
)

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Роли', to: '/admin/security/roles' },
  { label: currentRole.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="rolesLoading && !currentRole" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentRole">
      <PageHeader :title="currentRole.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <Button
            variant="destructive"
            size="sm"
            @click="showDeleteDialog = true"
          >
            Удалить роль
          </Button>
        </template>
      </PageHeader>

      <ErrorAlert v-if="rolesError" :message="rolesError" class="mb-4" />

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
              <Label for="parentId">Родительская роль</Label>
              <Select :model-value="state.parentId ?? '__none__'" @update:model-value="onParentChange">
                <SelectTrigger>
                  <SelectValue placeholder="Без родителя" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">Без родителя</SelectItem>
                  <SelectItem v-for="role in availableParents" :key="role.id" :value="role.id">
                    {{ role.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="description">Описание</Label>
              <Textarea id="description" v-model="state.description" rows="3" />
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2">
          <Button type="submit" :disabled="rolesLoading">
            Сохранить
          </Button>
          <Button variant="outline" type="button" @click="router.back()">
            Отмена
          </Button>
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить роль?"
        :description="`Роль «${currentRole.label}» (${currentRole.apiName}) будет удалена без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteRole"
      />
    </template>
  </div>
</template>
