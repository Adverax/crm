<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { credentialsApi } from '@/api/credentials'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ActiveStatusBadge from '@/components/admin/ActiveStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X, Zap, ZapOff, Wifi } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Credential, UsageLogEntry } from '@/types/credentials'

const props = defineProps<{
  credentialId: string
}>()

const router = useRouter()
const toast = useToast()

const cred = ref<Credential | null>(null)
const loading = ref(false)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const usageLog = ref<UsageLogEntry[]>([])
const activeTab = ref('settings')

const form = ref({
  name: '',
  description: '',
  baseUrl: '',
})

async function loadCredential() {
  loading.value = true
  try {
    const response = await credentialsApi.get(props.credentialId)
    cred.value = response.data
    form.value = {
      name: response.data.name,
      description: response.data.description ?? '',
      baseUrl: response.data.baseUrl,
    }
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

async function loadUsageLog() {
  try {
    const response = await credentialsApi.getUsageLog(props.credentialId)
    usageLog.value = response.data ?? []
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(() => {
  loadCredential()
  loadUsageLog()
})
watch(() => props.credentialId, () => {
  loadCredential()
  loadUsageLog()
})

async function onSave() {
  submitting.value = true
  try {
    await credentialsApi.update(props.credentialId, {
      name: form.value.name,
      description: form.value.description,
      baseUrl: form.value.baseUrl,
    })
    toast.success('Credential updated')
    await loadCredential()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onTestConnection() {
  submitting.value = true
  try {
    await credentialsApi.testConnection(props.credentialId)
    toast.success('Connection successful')
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onToggleActive() {
  if (!cred.value) return
  submitting.value = true
  try {
    if (cred.value.isActive) {
      await credentialsApi.deactivate(props.credentialId)
      toast.success('Credential deactivated')
    } else {
      await credentialsApi.activate(props.credentialId)
      toast.success('Credential activated')
    }
    await loadCredential()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

async function onDelete() {
  try {
    await credentialsApi.delete(props.credentialId)
    toast.success('Credential deleted')
    router.push({ name: 'admin-credentials' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Credentials', to: '/admin/metadata/credentials' },
  { label: cred.value?.code ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="loading && !cred" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="cred">
      <PageHeader :title="cred.code" :breadcrumbs="breadcrumbs">
        <template #actions>
          <div class="flex items-center gap-2">
            <ActiveStatusBadge :is-active="cred.isActive" />
            <Badge variant="outline">{{ cred.type }}</Badge>
            <IconButton
              :icon="Wifi"
              tooltip="Test connection"
              variant="outline"
              data-testid="test-connection-btn"
              @click="onTestConnection"
            />
            <IconButton
              :icon="cred.isActive ? ZapOff : Zap"
              :tooltip="cred.isActive ? 'Deactivate' : 'Activate'"
              variant="outline"
              data-testid="toggle-active-btn"
              @click="onToggleActive"
            />
            <IconButton
              :icon="Trash2"
              tooltip="Delete credential"
              variant="destructive"
              data-testid="delete-credential-btn"
              @click="showDeleteDialog = true"
            />
          </div>
        </template>
      </PageHeader>

      <Tabs v-model="activeTab" class="mt-4">
        <TabsList>
          <TabsTrigger value="settings" data-testid="tab-settings">Settings</TabsTrigger>
          <TabsTrigger value="usage" data-testid="tab-usage">Usage Log</TabsTrigger>
        </TabsList>

        <TabsContent value="settings" class="mt-4">
          <form class="max-w-xl space-y-4" @submit.prevent="onSave">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <div class="space-y-2">
                  <Label>Code</Label>
                  <Input :model-value="cred.code" disabled class="font-mono" />
                </div>
                <div class="space-y-2">
                  <Label for="name">Name</Label>
                  <Input id="name" v-model="form.name" required data-testid="field-name" />
                </div>
                <div class="space-y-2">
                  <Label for="base-url">Base URL</Label>
                  <Input id="base-url" v-model="form.baseUrl" required data-testid="field-base-url" />
                </div>
                <div class="space-y-2">
                  <Label for="description">Description</Label>
                  <Textarea id="description" v-model="form.description" rows="2" data-testid="field-description" />
                </div>
              </CardContent>
            </Card>
            <div class="flex gap-2 items-center">
              <Button type="submit" :disabled="submitting" data-testid="save-btn">
                Save
              </Button>
              <IconButton :icon="X" tooltip="Cancel" variant="outline" @click="router.push({ name: 'admin-credentials' })" />
            </div>
          </form>
        </TabsContent>

        <TabsContent value="usage" class="mt-4">
          <Card v-if="usageLog.length === 0">
            <CardContent class="py-6 text-center text-muted-foreground">
              No usage log entries.
            </CardContent>
          </Card>
          <div v-else class="space-y-2">
            <Card v-for="entry in usageLog" :key="entry.id" data-testid="usage-row">
              <CardContent class="py-3 flex items-center justify-between">
                <div>
                  <span class="font-mono text-sm">{{ entry.requestUrl }}</span>
                  <span v-if="entry.procedureCode" class="text-xs text-muted-foreground ml-2">
                    via {{ entry.procedureCode }}
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <Badge :variant="entry.success ? 'default' : 'destructive'">
                    {{ entry.responseStatus ?? 'N/A' }}
                  </Badge>
                  <span class="text-xs text-muted-foreground">{{ entry.durationMs }}ms</span>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete credential?"
        :description="`Credential '${cred.code}' will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
