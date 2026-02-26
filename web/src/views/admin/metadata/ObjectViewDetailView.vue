<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { objectViewsApi } from '@/api/object-views'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X, Eye, Pencil } from 'lucide-vue-next'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import type {
  ObjectView,
  OVAction,
  OVQuery,
  OVReadComputed,
  OVMutation,
  OVValidation,
  OVDefault,
  OVComputed,
} from '@/types/object-views'
import OVGeneralTab from '@/components/admin/object-view/OVGeneralTab.vue'
import OVFieldsTab from '@/components/admin/object-view/OVFieldsTab.vue'
import OVActionsTab from '@/components/admin/object-view/OVActionsTab.vue'
import OVQueriesTab from '@/components/admin/object-view/OVQueriesTab.vue'
import OVReadComputedTab from '@/components/admin/object-view/OVReadComputedTab.vue'
import OVMutationsTab from '@/components/admin/object-view/OVMutationsTab.vue'
import OVValidationTab from '@/components/admin/object-view/OVValidationTab.vue'
import OVDefaultsTab from '@/components/admin/object-view/OVDefaultsTab.vue'
import OVComputedTab from '@/components/admin/object-view/OVComputedTab.vue'

interface FormConfig {
  read: {
    fields: string[]
    actions: OVAction[]
    queries: OVQuery[]
    computed: OVReadComputed[]
  }
  write?: {
    fields?: string[]
    validation: OVValidation[]
    defaults: OVDefault[]
    computed: OVComputed[]
    mutations: OVMutation[]
  }
}

const props = defineProps<{
  viewId: string
}>()

const router = useRouter()
const toast = useToast()

const view = ref<ObjectView | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const error = ref<string | null>(null)

function emptyWriteConfig(): Required<FormConfig>['write'] {
  return {
    validation: [],
    defaults: [],
    computed: [],
    mutations: [],
  }
}

const form = ref<{ label: string; description: string; config: FormConfig }>({
  label: '',
  description: '',
  config: {
    read: {
      fields: [],
      actions: [],
      queries: [],
      computed: [],
    },
  },
})

function ensureWrite(): Required<FormConfig>['write'] {
  if (!form.value.config.write) {
    form.value.config.write = emptyWriteConfig()
  }
  return form.value.config.write
}

async function loadView() {
  loading.value = true
  error.value = null
  try {
    const response = await objectViewsApi.get(props.viewId)
    view.value = response.data
    const cfg = response.data.config
    form.value = {
      label: response.data.label ?? '',
      description: response.data.description ?? '',
      config: {
        read: {
          fields: cfg?.read?.fields ?? [],
          actions: cfg?.read?.actions ?? [],
          queries: cfg?.read?.queries ?? [],
          computed: cfg?.read?.computed ?? [],
        },
        write: cfg?.write ? {
          fields: cfg.write.fields,
          validation: cfg.write.validation ?? [],
          defaults: cfg.write.defaults ?? [],
          computed: cfg.write.computed ?? [],
          mutations: cfg.write.mutations ?? [],
        } : undefined,
      },
    }
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load object view: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadView)
watch(() => props.viewId, loadView)

async function onSave() {
  submitting.value = true
  try {
    await objectViewsApi.update(props.viewId, {
      label: form.value.label,
      description: form.value.description || undefined,
      config: form.value.config,
    })
    toast.success('Object view updated')
    router.push({ name: 'admin-object-views' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await objectViewsApi.delete(props.viewId)
    toast.success('Object view deleted')
    router.push({ name: 'admin-object-views' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-object-views' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Object Views', to: '/admin/metadata/object-views' },
  { label: view.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="loading && !view" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <ErrorAlert v-else-if="error" :message="error" class="mb-4" />

    <template v-else-if="view">
      <PageHeader :title="view.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <IconButton
            :icon="Trash2"
            tooltip="Delete object view"
            variant="destructive"
            data-testid="delete-view-btn"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <Tabs default-value="general" class="mt-4">
        <div class="flex items-center gap-2 mb-1">
          <Eye class="h-4 w-4 text-muted-foreground" />
          <span class="text-xs font-medium text-muted-foreground uppercase tracking-wide">Read</span>
        </div>
        <TabsList data-testid="view-tabs">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="fields">Fields</TabsTrigger>
          <TabsTrigger value="actions">Actions</TabsTrigger>
          <TabsTrigger value="queries">Queries</TabsTrigger>
          <TabsTrigger value="read-computed">Computed</TabsTrigger>
        </TabsList>

        <div class="flex items-center gap-2 mb-1 mt-3">
          <Pencil class="h-4 w-4 text-muted-foreground" />
          <span class="text-xs font-medium text-muted-foreground uppercase tracking-wide">Write</span>
        </div>
        <TabsList data-testid="data-tabs">
          <TabsTrigger value="validation">Validation</TabsTrigger>
          <TabsTrigger value="defaults">Defaults</TabsTrigger>
          <TabsTrigger value="write-computed">Computed</TabsTrigger>
          <TabsTrigger value="mutations">Mutations</TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <OVGeneralTab
            :view="view"
            :form="form"
            @update:label="form.label = $event"
            @update:description="form.description = $event"
          />
        </TabsContent>

        <TabsContent value="fields">
          <OVFieldsTab
            :fields="form.config.read.fields"
            @update:fields="form.config.read.fields = $event"
          />
        </TabsContent>

        <TabsContent value="actions">
          <OVActionsTab
            :actions="form.config.read.actions"
            @update:actions="form.config.read.actions = $event"
          />
        </TabsContent>

        <TabsContent value="queries">
          <OVQueriesTab
            :queries="form.config.read.queries"
            @update:queries="form.config.read.queries = $event"
          />
        </TabsContent>

        <TabsContent value="read-computed">
          <OVReadComputedTab
            :computed="form.config.read.computed"
            @update:computed="form.config.read.computed = $event"
          />
        </TabsContent>

        <TabsContent value="mutations">
          <OVMutationsTab
            :mutations="ensureWrite().mutations"
            @update:mutations="ensureWrite().mutations = $event"
          />
        </TabsContent>

        <TabsContent value="validation">
          <OVValidationTab
            :validation="ensureWrite().validation"
            @update:validation="ensureWrite().validation = $event"
          />
        </TabsContent>

        <TabsContent value="defaults">
          <OVDefaultsTab
            :defaults="ensureWrite().defaults"
            @update:defaults="ensureWrite().defaults = $event"
          />
        </TabsContent>

        <TabsContent value="write-computed">
          <OVComputedTab
            :computed="ensureWrite().computed"
            @update:computed="ensureWrite().computed = $event"
          />
        </TabsContent>
      </Tabs>

      <Separator class="my-6" />

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

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete object view?"
        :description="`Object view '${view.label}' will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
