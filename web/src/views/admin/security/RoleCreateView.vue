<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useRoleForm } from '@/composables/useRoleForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
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
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { rolesLoading, rolesError, roles } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useRoleForm()

onMounted(() => {
  store.fetchRoles({ perPage: 1000 }).catch((err) => toast.errorFromApi(err))
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onParentChange(value: any) {
  state.parentId = value === '__none__' ? null : String(value)
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createRole(toCreateRequest())
    toast.success('Роль создана')
    router.push({ name: 'admin-role-detail', params: { roleId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Роли', to: '/admin/security/roles' },
  { label: 'Новая роль' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать роль" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="rolesError" :message="rolesError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="sales_manager"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Название</Label>
            <Input id="label" v-model="state.label" placeholder="Менеджер по продажам" />
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
                <SelectItem v-for="role in roles" :key="role.id" :value="role.id">
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
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
