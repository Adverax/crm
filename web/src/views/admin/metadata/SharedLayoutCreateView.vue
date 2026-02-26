<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { sharedLayoutsApi } from '@/api/shared-layouts'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const SHARED_LAYOUT_TYPES = ['field', 'section', 'list'] as const

const router = useRouter()
const toast = useToast()
const submitting = ref(false)
const configJsonError = ref<string | null>(null)

const form = ref({
  apiName: '',
  type: 'field',
  label: '',
  configJson: '{}',
})

function validateConfigJson(): unknown | null {
  configJsonError.value = null
  try {
    return JSON.parse(form.value.configJson)
  } catch (err) {
    configJsonError.value = err instanceof Error ? err.message : 'Invalid JSON'
    return null
  }
}

async function onSubmit() {
  const config = validateConfigJson()
  if (config === null) {
    toast.error('Invalid JSON in config')
    return
  }

  submitting.value = true
  try {
    const result = await sharedLayoutsApi.create({
      apiName: form.value.apiName,
      type: form.value.type,
      label: form.value.label,
      config,
    })
    toast.success('Shared layout created')
    await router.push({
      name: 'admin-shared-layout-detail',
      params: { sharedLayoutId: result.data.id },
    })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-shared-layouts' })
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onTypeChange(value: any) {
  form.value.type = String(value)
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Shared Layouts', to: '/admin/metadata/shared-layouts' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Shared Layout" :breadcrumbs="breadcrumbs" />

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
                placeholder="compact_address_fields"
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
                placeholder="Compact Address Fields"
                data-testid="field-label"
              />
            </div>
          </div>

          <div class="space-y-2">
            <Label>Type</Label>
            <Select :model-value="form.type" @update:model-value="onTypeChange">
              <SelectTrigger data-testid="field-type">
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="t in SHARED_LAYOUT_TYPES"
                  :key="t"
                  :value="t"
                >
                  {{ t }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="config-json">Config (JSON)</Label>
            <Textarea
              id="config-json"
              v-model="form.configJson"
              rows="10"
              class="font-mono text-sm"
              data-testid="field-config"
            />
            <p v-if="configJsonError" class="text-sm text-destructive" data-testid="config-error">
              {{ configJsonError }}
            </p>
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
