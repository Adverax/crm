<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { layoutsApi } from '@/api/layouts'
import { objectViewsApi } from '@/api/object-views'
import { sharedLayoutsApi } from '@/api/shared-layouts'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import FormLayoutTab from '@/components/admin/layouts/FormLayoutTab.vue'
import ListConfigTab from '@/components/admin/layouts/ListConfigTab.vue'
import JsonTab from '@/components/admin/layouts/JsonTab.vue'
import type { Layout, LayoutConfig, SectionConfig, LayoutFieldConfig, ListConfig, SharedLayout } from '@/types/layouts'
import type { ObjectView } from '@/types/object-views'
import type { OVSection } from '@/components/admin/layouts/FormLayoutTab.vue'
import type { SectionField } from '@/components/admin/layouts/SectionCard.vue'

const props = defineProps<{
  layoutId: string
}>()

const router = useRouter()
const toast = useToast()

const layout = ref<Layout | null>(null)
const ov = ref<ObjectView | null>(null)
const sharedLayouts = ref<SharedLayout[]>([])
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const error = ref<string | null>(null)

// Reactive config form (separate parts for each tab)
const sectionConfig = ref<Record<string, SectionConfig>>({})
const fieldConfig = ref<Record<string, LayoutFieldConfig>>({})
const listConfig = ref<ListConfig>({})
const rootConfig = ref<unknown>(null)

async function loadLayout() {
  loading.value = true
  error.value = null
  try {
    const [layoutRes, sharedRes] = await Promise.all([
      layoutsApi.get(props.layoutId),
      sharedLayoutsApi.list(),
    ])
    layout.value = layoutRes.data
    sharedLayouts.value = sharedRes.data ?? []

    // Initialize config form from layout
    const cfg = layoutRes.data.config ?? {}
    sectionConfig.value = { ...cfg.sectionConfig }
    fieldConfig.value = { ...cfg.fieldConfig }
    listConfig.value = { ...cfg.listConfig }
    rootConfig.value = cfg.root ?? null

    // Load OV for section/field metadata
    try {
      const ovRes = await objectViewsApi.get(layoutRes.data.objectViewId)
      ov.value = ovRes.data
    } catch {
      ov.value = null
    }
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load layout: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadLayout)
watch(() => props.layoutId, loadLayout)

// Build the full config object from separate refs
const fullConfig = computed<LayoutConfig>(() => ({
  root: rootConfig.value as LayoutConfig['root'],
  sectionConfig: Object.keys(sectionConfig.value).length ? sectionConfig.value : undefined,
  fieldConfig: Object.keys(fieldConfig.value).length ? fieldConfig.value : undefined,
  listConfig: Object.keys(listConfig.value).length ? listConfig.value : undefined,
}))

// Derive OV sections for the canvas
const ovSections = computed<OVSection[]>(() => {
  const ovFields = ov.value?.config?.view?.fields ?? []
  if (!ovFields.length) return []

  // Extract section keys from root component tree
  const rootSections = extractSectionKeys(layout.value?.config?.root)

  // If root has field_section components, use those as sections
  if (rootSections.length > 0) {
    return rootSections.map((key) => ({
      key,
      label: key.charAt(0).toUpperCase() + key.slice(1),
      fields: ovFields.map(fieldToSectionField),
    }))
  }

  // Default: single "details" section with all OV fields
  return [{
    key: 'details',
    label: 'Details',
    fields: ovFields.map(fieldToSectionField),
  }]
})

// All OV fields as flat list (for ListConfigTab)
const allFields = computed<SectionField[]>(() => {
  const ovFields = ov.value?.config?.view?.fields ?? []
  return ovFields.map(fieldToSectionField)
})

function fieldToSectionField(fieldName: string): SectionField {
  return {
    apiName: fieldName,
    label: fieldName,
    type: 'text',
  }
}

function extractSectionKeys(root: unknown): string[] {
  if (!root || typeof root !== 'object') return []
  const node = root as Record<string, unknown>
  const keys: string[] = []
  if (node.type === 'field_section' && typeof node.key === 'string') {
    keys.push(node.key)
  }
  if (Array.isArray(node.children)) {
    for (const child of node.children) {
      keys.push(...extractSectionKeys(child))
    }
  }
  return keys
}

// JSON tab sync: when JSON is edited, update all refs
function onJsonConfigUpdate(config: LayoutConfig) {
  sectionConfig.value = { ...config.sectionConfig }
  fieldConfig.value = { ...config.fieldConfig }
  listConfig.value = { ...config.listConfig }
  rootConfig.value = config.root ?? null
}

async function onSave() {
  submitting.value = true
  try {
    await layoutsApi.update(props.layoutId, { config: fullConfig.value })
    toast.success('Layout updated')
    router.push({ name: 'admin-layouts' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await layoutsApi.delete(props.layoutId)
    toast.success('Layout deleted')
    router.push({ name: 'admin-layouts' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-layouts' })
}

const ovLabel = computed(() => ov.value?.label ?? layout.value?.objectViewId ?? '')

const heading = computed(() => {
  if (!layout.value) return '...'
  return `${ovLabel.value} / ${layout.value.formFactor} / ${layout.value.mode}`
})

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Layouts', to: '/admin/metadata/layouts' },
  { label: heading.value },
])
</script>

<template>
  <div>
    <div v-if="loading && !layout" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <ErrorAlert v-else-if="error" :message="error" class="mb-4" />

    <template v-else-if="layout">
      <PageHeader :title="heading" :breadcrumbs="breadcrumbs">
        <template #actions>
          <IconButton
            :icon="Trash2"
            tooltip="Delete layout"
            variant="destructive"
            data-testid="delete-layout-btn"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <div class="mt-4 space-y-4">
        <!-- Info badges -->
        <div class="flex items-center gap-4 text-sm">
          <div class="flex items-center gap-2">
            <span class="text-muted-foreground">Object View:</span>
            <span class="font-medium">{{ ovLabel }}</span>
          </div>
          <Badge data-testid="badge-form-factor">{{ layout.formFactor }}</Badge>
          <Badge variant="outline" data-testid="badge-mode">{{ layout.mode }}</Badge>
        </div>

        <!-- Tabs -->
        <Tabs default-value="form-layout" class="w-full">
          <TabsList data-testid="layout-tabs">
            <TabsTrigger value="form-layout" data-testid="tab-form-layout">Form Layout</TabsTrigger>
            <TabsTrigger value="list-config" data-testid="tab-list-config">List Config</TabsTrigger>
            <TabsTrigger value="json" data-testid="tab-json">JSON</TabsTrigger>
          </TabsList>

          <TabsContent value="form-layout" class="mt-4">
            <FormLayoutTab
              :sections="ovSections"
              :section-config="sectionConfig"
              :field-config="fieldConfig"
              :shared-layouts="sharedLayouts"
              @update:section-config="sectionConfig = $event"
              @update:field-config="fieldConfig = $event"
            />
          </TabsContent>

          <TabsContent value="list-config" class="mt-4">
            <ListConfigTab
              :list-config="listConfig"
              :available-fields="allFields"
              @update:list-config="listConfig = $event"
            />
          </TabsContent>

          <TabsContent value="json" class="mt-4">
            <JsonTab
              :config="fullConfig"
              @update:config="onJsonConfigUpdate"
            />
          </TabsContent>
        </Tabs>

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
        title="Delete layout?"
        :description="`Layout for '${ovLabel}' (${layout.formFactor}/${layout.mode}) will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
