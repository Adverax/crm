<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { validationRulesApi } from '@/api/validationRules'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { ValidationRule } from '@/types/validationRules'

const props = defineProps<{
  objectId: string
}>()

const router = useRouter()
const toast = useToast()

const rules = ref<ValidationRule[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadRules() {
  loading.value = true
  error.value = null
  try {
    const response = await validationRulesApi.list(props.objectId)
    rules.value = response.data ?? []
  } catch (err) {
    error.value = 'Не удалось загрузить правила валидации'
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
  { label: 'Админ', to: '/admin' },
  { label: 'Объекты', to: '/admin/metadata/objects' },
  { label: 'Объект', to: `/admin/metadata/objects/${props.objectId}` },
  { label: 'Правила валидации' },
])
</script>

<template>
  <div>
    <PageHeader title="Правила валидации" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button size="sm" data-testid="create-rule-btn" @click="goToCreate">
          Создать правило
        </Button>
      </template>
    </PageHeader>

    <ErrorAlert v-if="error" :message="error" class="mb-4" />

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="rules.length === 0"
      title="Нет правил валидации"
      description="Создайте первое правило валидации для этого объекта."
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
              Неактивно
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
