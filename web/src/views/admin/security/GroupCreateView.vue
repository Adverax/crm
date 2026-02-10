<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useGroupForm } from '@/composables/useGroupForm'
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
import type { GroupType } from '@/types/security'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { groupsLoading, groupsError } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useGroupForm()

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onGroupTypeChange(value: any) {
  state.groupType = String(value) as GroupType
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createGroup(toCreateRequest())
    toast.success('Группа создана')
    router.push({ name: 'admin-group-detail', params: { groupId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Группы', to: '/admin/security/groups' },
  { label: 'Новая группа' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать группу" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="groupsError" :message="groupsError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="sales_team"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Название</Label>
            <Input id="label" v-model="state.label" placeholder="Отдел продаж" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="groupType">Тип группы</Label>
            <Select :model-value="state.groupType" @update:model-value="onGroupTypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="public">Публичная</SelectItem>
                <SelectItem value="personal">Персональная</SelectItem>
                <SelectItem value="role">Роль</SelectItem>
                <SelectItem value="role_and_subordinates">Роль и подчинённые</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="groupsLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
