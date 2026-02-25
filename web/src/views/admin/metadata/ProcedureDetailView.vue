<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { proceduresApi } from '@/api/procedures'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ProcedureConstructor from '@/components/admin/procedures/ProcedureConstructor.vue'
import DryRunPanel from '@/components/admin/procedures/DryRunPanel.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Save, RotateCcw, Upload } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { ProcedureWithVersions, ProcedureVersion, ProcedureDefinition, CommandDef } from '@/types/procedures'

const props = defineProps<{
  procedureId: string
}>()

const router = useRouter()
const toast = useToast()

const data = ref<ProcedureWithVersions | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const versions = ref<ProcedureVersion[]>([])
const activeTab = ref('definition')

const form = ref({
  name: '',
  description: '',
})

const definition = ref<ProcedureDefinition>({
  commands: [],
  result: {},
})

async function loadProcedure() {
  loading.value = true
  try {
    const response = await proceduresApi.get(props.procedureId)
    data.value = response.data
    form.value = {
      name: response.data.procedure.name,
      description: response.data.procedure.description ?? '',
    }
    // Load draft definition if exists, otherwise published
    if (response.data.draftVersion) {
      definition.value = { ...response.data.draftVersion.definition }
    } else if (response.data.publishedVersion) {
      definition.value = { ...response.data.publishedVersion.definition }
    }
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

async function loadVersions() {
  try {
    const response = await proceduresApi.listVersions(props.procedureId)
    versions.value = response.data ?? []
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(() => {
  loadProcedure()
  loadVersions()
})
watch(() => props.procedureId, () => {
  loadProcedure()
  loadVersions()
})

async function onSaveDraft() {
  submitting.value = true
  try {
    await proceduresApi.saveDraft(props.procedureId, {
      definition: definition.value,
      changeSummary: 'Updated via constructor',
    })
    toast.success('Draft saved')
    await loadProcedure()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onPublish() {
  submitting.value = true
  try {
    await proceduresApi.publish(props.procedureId)
    toast.success('Procedure published')
    await loadProcedure()
    await loadVersions()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onRollback() {
  submitting.value = true
  try {
    await proceduresApi.rollback(props.procedureId)
    toast.success('Rolled back to previous version')
    await loadProcedure()
    await loadVersions()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onUpdateMetadata() {
  submitting.value = true
  try {
    await proceduresApi.updateMetadata(props.procedureId, {
      name: form.value.name,
      description: form.value.description,
    })
    toast.success('Settings updated')
    await loadProcedure()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await proceduresApi.delete(props.procedureId)
    toast.success('Procedure deleted')
    router.push({ name: 'admin-procedures' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function onCommandsUpdate(commands: CommandDef[]) {
  definition.value = { ...definition.value, commands }
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Procedures', to: '/admin/metadata/procedures' },
  { label: data.value?.procedure.code ?? '...' },
])

const hasDraft = computed(() => !!data.value?.procedure.draftVersionId)
const hasPublished = computed(() => !!data.value?.procedure.publishedVersionId)
</script>

<template>
  <div>
    <div v-if="loading && !data" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="data">
      <PageHeader :title="data.procedure.code" :breadcrumbs="breadcrumbs">
        <template #actions>
          <div class="flex items-center gap-2">
            <Badge v-if="hasDraft" variant="secondary">Draft</Badge>
            <Badge v-if="hasPublished" variant="default">Published</Badge>
            <IconButton
              :icon="Save"
              tooltip="Save draft"
              variant="outline"
              data-testid="save-draft-btn"
              @click="onSaveDraft"
            />
            <IconButton
              v-if="hasDraft"
              :icon="Upload"
              tooltip="Publish"
              variant="default"
              data-testid="publish-btn"
              @click="onPublish"
            />
            <IconButton
              v-if="hasPublished"
              :icon="RotateCcw"
              tooltip="Rollback"
              variant="outline"
              data-testid="rollback-btn"
              @click="onRollback"
            />
            <IconButton
              :icon="Trash2"
              tooltip="Delete procedure"
              variant="destructive"
              data-testid="delete-procedure-btn"
              @click="showDeleteDialog = true"
            />
          </div>
        </template>
      </PageHeader>

      <Tabs v-model="activeTab" class="mt-4">
        <TabsList>
          <TabsTrigger value="definition" data-testid="tab-definition">Definition</TabsTrigger>
          <TabsTrigger value="versions" data-testid="tab-versions">Versions</TabsTrigger>
          <TabsTrigger value="settings" data-testid="tab-settings">Settings</TabsTrigger>
          <TabsTrigger value="dry-run" data-testid="tab-dry-run">Dry Run</TabsTrigger>
        </TabsList>

        <TabsContent value="definition" class="mt-4">
          <ProcedureConstructor
            :commands="definition.commands"
            @update:commands="onCommandsUpdate"
          />
        </TabsContent>

        <TabsContent value="versions" class="mt-4">
          <Card v-if="versions.length === 0">
            <CardContent class="py-6 text-center text-muted-foreground">
              No versions yet.
            </CardContent>
          </Card>
          <div v-else class="space-y-2">
            <Card v-for="v in versions" :key="v.id" data-testid="version-row">
              <CardContent class="py-3 flex items-center justify-between">
                <div>
                  <span class="font-mono font-medium">v{{ v.version }}</span>
                  <span class="text-sm text-muted-foreground ml-2">
                    {{ v.changeSummary || 'No summary' }}
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <Badge
                    :variant="v.status === 'published' ? 'default' : v.status === 'draft' ? 'secondary' : 'outline'"
                  >
                    {{ v.status }}
                  </Badge>
                  <span class="text-xs text-muted-foreground">
                    {{ new Date(v.createdAt).toLocaleDateString() }}
                  </span>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="settings" class="mt-4">
          <form class="max-w-xl space-y-4" @submit.prevent="onUpdateMetadata">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <div class="space-y-2">
                  <Label>Code</Label>
                  <Input :model-value="data.procedure.code" disabled class="font-mono" />
                </div>
                <div class="space-y-2">
                  <Label for="name">Name</Label>
                  <Input id="name" v-model="form.name" required data-testid="field-name" />
                </div>
                <div class="space-y-2">
                  <Label for="description">Description</Label>
                  <Textarea id="description" v-model="form.description" rows="2" data-testid="field-description" />
                </div>
              </CardContent>
            </Card>
            <Button type="submit" :disabled="submitting" data-testid="save-settings-btn">
              Save Settings
            </Button>
          </form>
        </TabsContent>

        <TabsContent value="dry-run" class="mt-4">
          <DryRunPanel :procedure-id="procedureId" />
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete procedure?"
        :description="`Procedure '${data.procedure.code}' will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
