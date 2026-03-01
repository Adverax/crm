<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { portalsApi } from '@/api/portals'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2 } from 'lucide-vue-next'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import type {
  Portal,
  PortalAction,
  PortalQuery,
  PortalViewField,
} from '@/types/portals'
import PortalGeneralTab from '@/components/admin/portal/PortalGeneralTab.vue'
import PortalFieldsTab from '@/components/admin/portal/PortalFieldsTab.vue'
import PortalActionsTab from '@/components/admin/portal/PortalActionsTab.vue'
import PortalQueriesTab from '@/components/admin/portal/PortalQueriesTab.vue'

interface FormConfig {
  read: {
    fields: PortalViewField[]
    actions: PortalAction[]
    queries: PortalQuery[]
  }
}

const props = defineProps<{
  portalId: string
}>()

const router = useRouter()
const toast = useToast()

const view = ref<Portal | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const error = ref<string | null>(null)

const form = ref<{ label: string; description: string; config: FormConfig }>({
  label: '',
  description: '',
  config: {
    read: {
      fields: [],
      actions: [],
      queries: [],
    },
  },
})

async function loadView() {
  loading.value = true
  error.value = null
  try {
    const response = await portalsApi.get(props.portalId)
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
        },
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
watch(() => props.portalId, loadView)

async function onSave() {
  submitting.value = true
  try {
    await portalsApi.update(props.portalId, {
      label: form.value.label,
      description: form.value.description || undefined,
      config: form.value.config,
    })
    toast.success('Object view updated')
    router.push({ name: 'admin-portals' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await portalsApi.delete(props.portalId)
    toast.success('Object view deleted')
    router.push({ name: 'admin-portals' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Portals', to: '/admin/metadata/portals' },
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
          <Button type="button" :disabled="submitting" data-testid="save-btn" @click="onSave">
            Save
          </Button>
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
        <TabsList data-testid="view-tabs">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="queries">Queries</TabsTrigger>
          <TabsTrigger value="fields">Fields</TabsTrigger>
          <TabsTrigger value="actions">Actions</TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <PortalGeneralTab
            :view="view"
            :form="form"
            @update:label="form.label = $event"
            @update:description="form.description = $event"
          />
        </TabsContent>

        <TabsContent value="fields">
          <PortalFieldsTab
            :fields="form.config.read.fields"
            @update:fields="form.config.read.fields = $event"
          />
        </TabsContent>

        <TabsContent value="actions">
          <PortalActionsTab
            :actions="form.config.read.actions"
            @update:actions="form.config.read.actions = $event"
          />
        </TabsContent>

        <TabsContent value="queries">
          <PortalQueriesTab
            :queries="form.config.read.queries"
            @update:queries="form.config.read.queries = $event"
          />
        </TabsContent>

      </Tabs>

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
