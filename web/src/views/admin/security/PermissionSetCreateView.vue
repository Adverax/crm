<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { usePermissionSetForm } from '@/composables/usePermissionSetForm'
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
import type { PsType } from '@/types/security'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { isLoading, error } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = usePermissionSetForm()

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onPsTypeChange(value: any) {
  state.psType = String(value) as PsType
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createPermissionSet(toCreateRequest())
    toast.success('Набор разрешений создан')
    router.push({ name: 'admin-permission-set-detail', params: { permissionSetId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Наборы разрешений', to: '/admin/security/permission-sets' },
  { label: 'Новый набор' },
]
</script>

<template>
  <div>
    <PageHeader title="Создать набор разрешений" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">Основная информация</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="sales_read_access"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Название</Label>
            <Input id="label" v-model="state.label" placeholder="Доступ на чтение продаж" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="psType">Тип</Label>
            <Select :model-value="state.psType" @update:model-value="onPsTypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="grant">Grant (разрешает)</SelectItem>
                <SelectItem value="deny">Deny (запрещает)</SelectItem>
              </SelectContent>
            </Select>
            <p v-if="errors.psType" class="text-sm text-destructive">{{ errors.psType }}</p>
          </div>

          <div class="space-y-2">
            <Label for="description">Описание</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2">
        <Button type="submit" :disabled="isLoading">
          Создать
        </Button>
        <Button variant="outline" type="button" @click="router.back()">
          Отмена
        </Button>
      </div>
    </form>
  </div>
</template>
