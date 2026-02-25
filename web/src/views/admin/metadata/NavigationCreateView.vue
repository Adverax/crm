<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { navigationApi } from '@/api/navigation'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const profileId = ref('')
const configJson = ref(JSON.stringify({
  groups: [
    {
      key: 'main',
      label: 'Main',
      icon: 'briefcase',
      items: [
        { type: 'object', object_api_name: 'Account' },
        { type: 'object', object_api_name: 'Contact' },
      ],
    },
  ],
}, null, 2))

async function onSubmit() {
  submitting.value = true
  try {
    const config = JSON.parse(configJson.value)
    const response = await navigationApi.create({
      profile_id: profileId.value,
      config,
    })
    toast.success('Navigation created')
    router.push({ name: 'admin-navigation-detail', params: { navigationId: response.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-navigation' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Navigation', to: '/admin/metadata/navigation' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Navigation" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="space-y-2">
            <Label for="profile-id">Profile ID</Label>
            <Input id="profile-id" v-model="profileId" required placeholder="UUID of the profile" class="font-mono" data-testid="field-profile-id" />
          </div>
          <div class="space-y-2">
            <Label for="config">Config (JSON)</Label>
            <Textarea
              id="config"
              v-model="configJson"
              rows="16"
              class="font-mono text-sm"
              data-testid="field-config"
            />
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton :icon="X" tooltip="Cancel" variant="outline" data-testid="cancel-btn" @click="onCancel" />
      </div>
    </form>
  </div>
</template>
