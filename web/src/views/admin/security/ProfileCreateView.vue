<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useProfileForm } from '@/composables/useProfileForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { profilesLoading, profilesError } = storeToRefs(store)
const { state, errors, validate, toCreateRequest } = useProfileForm()

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await store.createProfile(toCreateRequest())
    toast.success('Profile created')
    router.push({ name: 'admin-profile-detail', params: { profileId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Profiles', to: '/admin/security/profiles' },
  { label: 'New Profile' },
]
</script>

<template>
  <div>
    <PageHeader title="Create Profile" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="profilesError" :message="profilesError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">General Information</h2>

          <div class="space-y-2">
            <Label for="apiName">API Name</Label>
            <Input
              id="apiName"
              v-model="state.apiName"
              placeholder="sales_profile"
            />
            <p v-if="errors.apiName" class="text-sm text-destructive">{{ errors.apiName }}</p>
          </div>

          <div class="space-y-2">
            <Label for="label">Label</Label>
            <Input id="label" v-model="state.label" placeholder="Sales Manager Profile" />
            <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
          </div>

          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea id="description" v-model="state.description" rows="3" />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="profilesLoading">
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
