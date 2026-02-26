<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { sharedLayoutsApi } from '@/api/shared-layouts'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { SharedLayout } from '@/types/layouts'

const props = defineProps<{
  sharedLayoutId: string
}>()

const router = useRouter()
const toast = useToast()

const sharedLayout = ref<SharedLayout | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const error = ref<string | null>(null)
const configJsonError = ref<string | null>(null)

const form = ref({
  label: '',
  configJson: '{}',
})

async function loadSharedLayout() {
  loading.value = true
  error.value = null
  try {
    const response = await sharedLayoutsApi.get(props.sharedLayoutId)
    sharedLayout.value = response.data
    form.value = {
      label: response.data.label,
      configJson: JSON.stringify(response.data.config ?? {}, null, 2),
    }
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load shared layout: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadSharedLayout)
watch(() => props.sharedLayoutId, loadSharedLayout)

function validateConfigJson(): unknown | null {
  configJsonError.value = null
  try {
    return JSON.parse(form.value.configJson)
  } catch (err) {
    configJsonError.value = err instanceof Error ? err.message : 'Invalid JSON'
    return null
  }
}

async function onSave() {
  const config = validateConfigJson()
  if (config === null) {
    toast.error('Invalid JSON in config')
    return
  }

  submitting.value = true
  try {
    await sharedLayoutsApi.update(props.sharedLayoutId, {
      label: form.value.label,
      config,
    })
    toast.success('Shared layout updated')
    router.push({ name: 'admin-shared-layouts' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await sharedLayoutsApi.delete(props.sharedLayoutId)
    toast.success('Shared layout deleted')
    router.push({ name: 'admin-shared-layouts' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-shared-layouts' })
}

function typeVariant(type: string): 'default' | 'secondary' | 'outline' {
  if (type === 'field') return 'default'
  if (type === 'section') return 'secondary'
  return 'outline'
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Shared Layouts', to: '/admin/metadata/shared-layouts' },
  { label: sharedLayout.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="loading && !sharedLayout" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <ErrorAlert v-else-if="error" :message="error" class="mb-4" />

    <template v-else-if="sharedLayout">
      <PageHeader :title="sharedLayout.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <IconButton
            :icon="Trash2"
            tooltip="Delete shared layout"
            variant="destructive"
            data-testid="delete-shared-layout-btn"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <div class="mt-4 max-w-4xl space-y-6">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="flex items-center gap-3">
              <span class="text-sm text-muted-foreground">API Name:</span>
              <span class="font-mono font-medium" data-testid="display-api-name">
                {{ sharedLayout.apiName }}
              </span>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-sm text-muted-foreground">Type:</span>
              <Badge :variant="typeVariant(sharedLayout.type)" data-testid="badge-type">
                {{ sharedLayout.type }}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent class="pt-6 space-y-4">
            <div class="space-y-2">
              <Label for="label">Label</Label>
              <Input
                id="label"
                v-model="form.label"
                required
                data-testid="field-label"
              />
            </div>

            <div class="space-y-2">
              <Label for="config-json">Config (JSON)</Label>
              <Textarea
                id="config-json"
                v-model="form.configJson"
                rows="20"
                class="font-mono text-sm"
                data-testid="field-config"
              />
              <p v-if="configJsonError" class="text-sm text-destructive" data-testid="config-error">
                {{ configJsonError }}
              </p>
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2 items-center">
          <Button type="button" :disabled="submitting" data-testid="save-btn" @click="onSave">
            Save
          </Button>
          <IconButton
            :icon="X"
            tooltip="Cancel"
            variant="outline"
            data-testid="cancel-btn"
            @click="onCancel"
          />
        </div>
      </div>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete shared layout?"
        :description="`Shared layout '${sharedLayout.label}' (${sharedLayout.apiName}) will be permanently deleted. Any layouts referencing it via layout_ref will lose the reference.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
