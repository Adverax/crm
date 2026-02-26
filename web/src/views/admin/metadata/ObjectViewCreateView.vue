<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { objectViewsApi } from '@/api/object-views'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface ProfileOption {
  id: string
  label: string
}

const NONE_PROFILE = '__none__'

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const profiles = ref<ProfileOption[]>([])

const form = ref({
  apiName: '',
  label: '',
  description: '',
  profileId: NONE_PROFILE,
  isDefault: false,
})

async function loadProfiles() {
  try {
    const { http } = await import('@/api/http')
    const response = await http.get<{ data: { id: string; label: string }[] }>('/api/v1/admin/security/profiles')
    profiles.value = (response.data ?? []).map((p) => ({
      id: p.id,
      label: p.label,
    }))
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(() => {
  loadProfiles()
})

async function onSubmit() {
  submitting.value = true
  try {
    const result = await objectViewsApi.create({
      profileId: form.value.profileId === NONE_PROFILE ? undefined : form.value.profileId,
      apiName: form.value.apiName,
      label: form.value.label,
      description: form.value.description || undefined,
      isDefault: form.value.isDefault,
      config: {
        read: {
          fields: [],
          actions: [],
          queries: [],
          computed: [],
        },
      },
    })
    toast.success('Object view created')
    await router.push({ name: 'admin-object-view-detail', params: { viewId: result.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-object-views' })
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onProfileChange(value: any) {
  form.value.profileId = String(value) || NONE_PROFILE
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Object Views', to: '/admin/metadata/object-views' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Object View" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="api-name">API Name</Label>
              <Input
                id="api-name"
                v-model="form.apiName"
                required
                placeholder="account_sales_view"
                class="font-mono"
                data-testid="field-api-name"
              />
              <p class="text-xs text-muted-foreground">
                Lowercase letters, numbers, and underscores
              </p>
            </div>
            <div class="space-y-2">
              <Label for="label">Label</Label>
              <Input
                id="label"
                v-model="form.label"
                required
                placeholder="Sales View"
                data-testid="field-label"
              />
            </div>
          </div>

          <div class="space-y-2">
            <Label>Profile</Label>
            <Select :model-value="form.profileId" @update:model-value="onProfileChange">
              <SelectTrigger data-testid="field-profile">
                <SelectValue placeholder="None (global)" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="NONE_PROFILE">None (global)</SelectItem>
                <SelectItem
                  v-for="profile in profiles"
                  :key="profile.id"
                  :value="profile.id"
                >
                  {{ profile.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea
              id="description"
              v-model="form.description"
              rows="2"
              data-testid="field-description"
            />
          </div>

          <div class="flex items-center gap-2">
            <Checkbox
              id="is-default"
              :checked="form.isDefault"
              data-testid="field-is-default"
              @update:checked="(v: boolean) => (form.isDefault = v)"
            />
            <Label for="is-default">Default view for this object</Label>
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton
          :icon="X"
          tooltip="Cancel"
          variant="outline"
          data-testid="cancel-btn"
          @click="onCancel"
        />
      </div>
    </form>
  </div>
</template>
