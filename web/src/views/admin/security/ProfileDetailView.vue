<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useProfileForm } from '@/composables/useProfileForm'
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
import { storeToRefs } from 'pinia'

const props = defineProps<{
  profileId: string
}>()

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { currentProfile, profilesLoading, profilesError } = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = useProfileForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const profile = await store.fetchProfile(props.profileId)
    initFrom(profile)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.profileId, loadData)

async function onSave() {
  if (!validate()) return
  try {
    await store.updateProfile(props.profileId, toUpdateRequest())
    toast.success('Профиль обновлён')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteProfile() {
  try {
    await store.deleteProfile(props.profileId)
    toast.success('Профиль удалён')
    router.push({ name: 'admin-profiles' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Профили', to: '/admin/security/profiles' },
  { label: currentProfile.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="profilesLoading && !currentProfile" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentProfile">
      <PageHeader :title="currentProfile.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <Button
            variant="destructive"
            size="sm"
            @click="showDeleteDialog = true"
          >
            Удалить профиль
          </Button>
        </template>
      </PageHeader>

      <ErrorAlert v-if="profilesError" :message="profilesError" class="mb-4" />

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

            <div class="space-y-2">
              <Label>Базовый набор разрешений</Label>
              <RouterLink
                :to="{ name: 'admin-permission-set-detail', params: { permissionSetId: currentProfile.basePermissionSetId } }"
                class="text-sm text-primary hover:underline block"
              >
                Открыть базовый набор разрешений
              </RouterLink>
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2">
          <Button type="submit" :disabled="profilesLoading">
            Сохранить
          </Button>
          <Button variant="outline" type="button" @click="router.back()">
            Отмена
          </Button>
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить профиль?"
        :description="`Профиль «${currentProfile.label}» (${currentProfile.apiName}) будет удалён без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteProfile"
      />
    </template>
  </div>
</template>
