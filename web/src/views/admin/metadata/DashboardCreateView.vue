<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { dashboardApi } from '@/api/dashboard'
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
  widgets: [
    {
      key: 'recent_tasks',
      type: 'list',
      label: 'My Recent Tasks',
      size: 'half',
      query: 'SELECT Id, subject, due_date FROM Task WHERE owner_id = :currentUserId LIMIT 5',
      columns: ['subject', 'due_date'],
      object_api_name: 'Task',
    },
  ],
}, null, 2))

async function onSubmit() {
  submitting.value = true
  try {
    const config = JSON.parse(configJson.value)
    const response = await dashboardApi.create({
      profile_id: profileId.value,
      config,
    })
    toast.success('Dashboard created')
    router.push({ name: 'admin-dashboard-detail', params: { dashboardId: response.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-dashboards' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Dashboards', to: '/admin/metadata/dashboards' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Dashboard" :breadcrumbs="breadcrumbs" />

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
