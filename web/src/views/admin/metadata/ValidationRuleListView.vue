<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { validationRulesApi } from '@/api/validationRules'
import { useToast } from '@/composables/useToast'
import { http } from '@/api/http'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { ValidationRule } from '@/types/validationRules'

const props = defineProps<{
  objectId: string
}>()

const router = useRouter()
const toast = useToast()

const objectLabel = ref('')
const rules = ref<ValidationRule[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadRules() {
  loading.value = true
  error.value = null
  try {
    const [rulesResp, objResp] = await Promise.all([
      validationRulesApi.list(props.objectId),
      http.get<{ data: { label: string } }>(`/api/v1/admin/metadata/objects/${props.objectId}`),
    ])
    rules.value = rulesResp.data ?? []
    objectLabel.value = objResp.data.label
  } catch (err) {
    error.value = 'Failed to load validation rules'
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

onMounted(loadRules)

function goToCreate() {
  router.push({
    name: 'admin-validation-rule-create',
    params: { objectId: props.objectId },
  })
}

function goToDetail(ruleId: string) {
  router.push({
    name: 'admin-validation-rule-detail',
    params: { objectId: props.objectId, ruleId },
  })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Objects', to: '/admin/metadata/objects' },
  { label: objectLabel.value || '...', to: `/admin/metadata/objects/${props.objectId}` },
  { label: 'Rules' },
])
</script>

<template>
  <div>
    <PageHeader title="Validation Rules" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create rule"
          variant="default"
          data-testid="create-rule-btn"
          @click="goToCreate"
        />
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="rules.length === 0"
      title="No validation rules"
      description="Create the first validation rule for this object."
    />

    <div v-else class="space-y-2">
      <Card
        v-for="rule in rules"
        :key="rule.id"
        class="cursor-pointer hover:bg-muted/50 transition-colors"
        data-testid="rule-row"
        @click="goToDetail(rule.id)"
      >
        <CardContent class="py-3 flex items-center justify-between">
          <div>
            <div class="font-medium">{{ rule.label }}</div>
            <div class="text-sm text-muted-foreground">{{ rule.apiName }}</div>
          </div>
          <div class="flex items-center gap-2">
            <Badge :variant="rule.severity === 'error' ? 'destructive' : 'secondary'">
              {{ rule.severity }}
            </Badge>
            <Badge v-if="!rule.isActive" variant="outline">
              Inactive
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
