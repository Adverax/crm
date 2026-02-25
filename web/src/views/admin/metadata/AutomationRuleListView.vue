<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { automationRulesApi } from '@/api/automation-rules'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ActiveStatusBadge from '@/components/admin/ActiveStatusBadge.vue'
import { Input } from '@/components/ui/input'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { AutomationRule } from '@/types/automation-rules'

interface ObjectOption {
  id: string
  label: string
}

const router = useRouter()
const toast = useToast()

const rules = ref<AutomationRule[]>([])
const objects = ref<ObjectOption[]>([])
const selectedObjectId = ref<string>('')
const searchQuery = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const filteredRules = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return rules.value
  return rules.value.filter(
    (rule) =>
      rule.name.toLowerCase().includes(q) ||
      rule.procedureCode.toLowerCase().includes(q),
  )
})

async function loadObjects() {
  try {
    const response = await fetch('/api/v1/admin/metadata/objects', {
      headers: { Authorization: `Bearer ${localStorage.getItem('crm_access_token')}` },
    })
    const data = await response.json()
    objects.value = (data.data ?? []).map((o: { id: string; label: string }) => ({
      id: o.id,
      label: o.label,
    }))
    if (objects.value.length > 0 && !selectedObjectId.value) {
      selectedObjectId.value = objects.value[0]!.id
      await loadRules()
    }
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function loadRules() {
  if (!selectedObjectId.value) return
  loading.value = true
  error.value = null
  try {
    const response = await automationRulesApi.list(selectedObjectId.value)
    rules.value = response.data ?? []
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err)
    error.value = `Failed to load automation rules: ${detail}`
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadObjects)

function goToCreate() {
  router.push({
    name: 'admin-automation-rule-create',
    params: { objectId: selectedObjectId.value },
  })
}

function goToDetail(id: string) {
  router.push({ name: 'admin-automation-rule-detail', params: { ruleId: id } })
}

function eventLabel(type: string): string {
  return type.replace('_', ' ')
}

const selectedObjectLabel = computed(() =>
  objects.value.find((o) => o.id === selectedObjectId.value)?.label ?? '',
)

const breadcrumbs = computed(() => {
  const crumbs: { label: string; to?: string }[] = [
    { label: 'Admin', to: '/admin' },
  ]
  if (selectedObjectLabel.value) {
    crumbs.push({ label: selectedObjectLabel.value })
  }
  crumbs.push({ label: 'Automation Rules' })
  return crumbs
})
</script>

<template>
  <div>
    <PageHeader title="Automation Rules" :breadcrumbs="breadcrumbs">
      <template #actions>
        <div class="flex items-center gap-2">
          <select
            v-model="selectedObjectId"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm"
            data-testid="object-select"
            @change="loadRules"
          >
            <option v-for="obj in objects" :key="obj.id" :value="obj.id">
              {{ obj.label }}
            </option>
          </select>
          <Input
            v-model="searchQuery"
            placeholder="Filter..."
            class="h-9 w-64"
            data-testid="search-input"
          />
          <IconButton
            :icon="Plus"
            tooltip="Create automation rule"
            variant="default"
            size="icon-sm"
            data-testid="create-rule-btn"
            @click="goToCreate"
          />
        </div>
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="filteredRules.length === 0"
      title="No automation rules"
      description="Create automation rules to react to data changes."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="rule in filteredRules"
        :key="rule.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="rule-row"
        @click="goToDetail(rule.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium">{{ rule.name }}</div>
            <div class="text-sm text-muted-foreground">
              {{ eventLabel(rule.eventType) }} &mdash; {{ rule.procedureCode }}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Badge variant="outline">{{ rule.executionMode }}</Badge>
            <ActiveStatusBadge :is-active="rule.isActive" />
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
