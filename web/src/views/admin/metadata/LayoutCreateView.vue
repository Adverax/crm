<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { layoutsApi } from '@/api/layouts'
import { objectViewsApi } from '@/api/object-views'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ObjectView } from '@/types/object-views'

interface OvOption {
  id: string
  label: string
}

const FORM_FACTORS = ['desktop', 'tablet', 'mobile'] as const
const MODES = ['read', 'view'] as const

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const objectViews = ref<OvOption[]>([])

const form = ref({
  objectViewId: '',
  formFactor: 'desktop',
  mode: 'read',
})

async function loadObjectViews() {
  try {
    const response = await objectViewsApi.list()
    objectViews.value = (response.data ?? []).map((ov: ObjectView) => ({
      id: ov.id!,
      label: ov.label,
    }))
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadObjectViews)

async function onSubmit() {
  if (!form.value.objectViewId) {
    toast.error('Please select an Object View')
    return
  }

  submitting.value = true
  try {
    const result = await layoutsApi.create({
      objectViewId: form.value.objectViewId,
      formFactor: form.value.formFactor,
      mode: form.value.mode,
      config: {},
    })
    toast.success('Layout created')
    await router.push({ name: 'admin-layout-detail', params: { layoutId: result.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-layouts' })
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onOvChange(value: any) {
  form.value.objectViewId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onFormFactorChange(value: any) {
  form.value.formFactor = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onModeChange(value: any) {
  form.value.mode = String(value)
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Layouts', to: '/admin/metadata/layouts' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Layout" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="space-y-2">
            <Label>Object View</Label>
            <Select :model-value="form.objectViewId" @update:model-value="onOvChange">
              <SelectTrigger data-testid="field-object-view">
                <SelectValue placeholder="Select an Object View" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="ov in objectViews"
                  :key="ov.id"
                  :value="ov.id"
                >
                  {{ ov.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label>Form Factor</Label>
              <Select :model-value="form.formFactor" @update:model-value="onFormFactorChange">
                <SelectTrigger data-testid="field-form-factor">
                  <SelectValue placeholder="Select form factor" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="ff in FORM_FACTORS"
                    :key="ff"
                    :value="ff"
                  >
                    {{ ff }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label>Mode</Label>
              <Select :model-value="form.mode" @update:model-value="onModeChange">
                <SelectTrigger data-testid="field-mode">
                  <SelectValue placeholder="Select mode" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="m in MODES"
                    :key="m"
                    :value="m"
                  >
                    {{ m }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
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
