<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { navigationApi } from '@/api/navigation'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X, Trash2 } from 'lucide-vue-next'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const props = defineProps<{
  navigationId: string
}>()

const router = useRouter()
const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const error = ref<string | null>(null)
const showDeleteDialog = ref(false)

const profileId = ref('')
const configJson = ref('')

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const response = await navigationApi.get(props.navigationId)
    const nav = response.data
    profileId.value = nav.profileId
    configJson.value = JSON.stringify(nav.config, null, 2)
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load navigation: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

async function onSave() {
  saving.value = true
  try {
    const config = JSON.parse(configJson.value)
    await navigationApi.update(props.navigationId, { config })
    toast.success('Navigation updated')
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    saving.value = false
  }
}

async function onDelete() {
  try {
    await navigationApi.delete(props.navigationId)
    toast.success('Navigation deleted')
    router.push({ name: 'admin-navigation' })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

function onCancel() {
  router.push({ name: 'admin-navigation' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Navigation', to: '/admin/metadata/navigation' },
  { label: profileId.value || 'Detail' },
])
</script>

<template>
  <div>
    <PageHeader title="Navigation Detail" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-4 mt-4 max-w-3xl">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-64 w-full" />
    </div>

    <form v-else class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSave">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="space-y-2">
            <Label>Profile ID</Label>
            <div class="font-mono text-sm text-muted-foreground" data-testid="profile-id">{{ profileId }}</div>
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
        <Button type="submit" :disabled="saving" data-testid="save-btn">
          Save
        </Button>
        <IconButton :icon="X" tooltip="Cancel" variant="outline" data-testid="cancel-btn" @click="onCancel" />
        <IconButton :icon="Trash2" tooltip="Delete" variant="destructive" data-testid="delete-btn" @click="showDeleteDialog = true" />
      </div>
    </form>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete Navigation"
      description="Are you sure you want to delete this navigation config? This action cannot be undone."
      @confirm="onDelete"
      @cancel="showDeleteDialog = false"
    />
  </div>
</template>
