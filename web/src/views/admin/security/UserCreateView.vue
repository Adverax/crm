<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useUserForm } from '@/composables/useUserForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
const { usersLoading, usersError, profiles, roles } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useUserForm()

onMounted(async () => {
  try {
    await Promise.all([
      store.fetchProfiles({ perPage: 1000 }),
      store.fetchRoles({ perPage: 1000 }),
    ])
  } catch (err) {
    toast.errorFromApi(err)
  }
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onProfileChange(value: any) {
  state.profileId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onRoleChange(value: any) {
  state.roleId = value === '__none__' ? null : String(value)
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createUser(toCreateRequest())
    toast.success('Пользователь создан')
    router.push({ name: 'admin-user-detail', params: { userId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Пользователи', to: '/admin/security/users' },
  { label: 'Новый пользователь' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать пользователя" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="usersError" :message="usersError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Учётные данные</h2>

          <div class="space-y-2">
            <Label for="username">Имя пользователя</Label>
            <Input
              id="username"
              v-model="state.username"
              placeholder="john.doe"
            />
            <p v-if="errors.username" class="text-sm text-destructive">{{ errors.username }}</p>
          </div>

          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input
              id="email"
              v-model="state.email"
              type="email"
              placeholder="john@example.com"
            />
            <p v-if="errors.email" class="text-sm text-destructive">{{ errors.email }}</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Личные данные</h2>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="firstName">Имя</Label>
              <Input id="firstName" v-model="state.firstName" placeholder="Иван" />
            </div>
            <div class="space-y-2">
              <Label for="lastName">Фамилия</Label>
              <Input id="lastName" v-model="state.lastName" placeholder="Иванов" />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Безопасность</h2>

          <div class="space-y-2">
            <Label for="profileId">Профиль</Label>
            <Select :model-value="state.profileId || undefined" @update:model-value="onProfileChange">
              <SelectTrigger>
                <SelectValue placeholder="Выберите профиль" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="profile in profiles" :key="profile.id" :value="profile.id">
                  {{ profile.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="errors.profileId" class="text-sm text-destructive">{{ errors.profileId }}</p>
          </div>

          <div class="space-y-2">
            <Label for="roleId">Роль</Label>
            <Select :model-value="state.roleId ?? '__none__'" @update:model-value="onRoleChange">
              <SelectTrigger>
                <SelectValue placeholder="Без роли" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">Без роли</SelectItem>
                <SelectItem v-for="role in roles" :key="role.id" :value="role.id">
                  {{ role.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="usersLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
