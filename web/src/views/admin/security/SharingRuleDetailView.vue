<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useMetadataStore } from '@/stores/metadata'
import { useSharingRuleForm } from '@/composables/useSharingRuleForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { storeToRefs } from 'pinia'
import type { AccessLevel } from '@/types/security'

const props = defineProps<{
  ruleId: string
}>()

const router = useRouter()
const securityStore = useSecurityAdminStore()
const metadataStore = useMetadataStore()
const toast = useToast()

const { currentSharingRule, sharingRulesLoading, sharingRulesError, groups } = storeToRefs(securityStore)
const { objects } = storeToRefs(metadataStore)
const { state, errors, validate, toUpdateRequest, initFrom } = useSharingRuleForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const rule = await securityStore.fetchSharingRule(props.ruleId)
    initFrom(rule)
    await metadataStore.fetchObjects({ perPage: 1000 })
    await securityStore.fetchGroups({ perPage: 1000 })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.ruleId, loadData)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onTargetGroupChange(value: any) {
  state.targetGroupId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onAccessLevelChange(value: any) {
  state.accessLevel = String(value) as AccessLevel
}

async function onSave() {
  if (!validate()) return
  try {
    await securityStore.updateSharingRule(props.ruleId, toUpdateRequest())
    toast.success('Rule updated')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteRule() {
  try {
    await securityStore.deleteSharingRule(props.ruleId)
    toast.success('Rule deleted')
    router.push({ name: 'admin-sharing-rules' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

function objectName(objectId: string): string {
  const obj = objects.value.find((o) => o.id === objectId)
  return obj?.label ?? objectId
}

function groupName(groupId: string): string {
  const group = groups.value.find((g) => g.id === groupId)
  return group?.label ?? groupId
}

function ruleTypeLabel(type: string): string {
  return type === 'owner_based' ? 'Owner-based' : 'Criteria-based'
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Sharing Rules', to: '/admin/security/sharing-rules' },
  { label: currentSharingRule.value ? objectName(currentSharingRule.value.objectId) : '...' },
])
</script>

<template>
  <div>
    <div v-if="sharingRulesLoading && !currentSharingRule" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentSharingRule">
      <PageHeader :title="`Rule: ${objectName(currentSharingRule.objectId)}`" :breadcrumbs="breadcrumbs">
        <template #actions>
          <IconButton
            :icon="Trash2"
            tooltip="Delete Rule"
            variant="destructive"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <ErrorAlert v-if="sharingRulesError" :message="sharingRulesError" class="mb-4" />

      <form class="max-w-2xl space-y-6" @submit.prevent="onSave">
        <Card>
          <CardContent class="pt-6 space-y-4">
            <h2 class="text-lg font-semibold">General Information</h2>

            <div class="space-y-2">
              <Label>Object</Label>
              <Input :model-value="objectName(currentSharingRule.objectId)" disabled />
            </div>

            <div class="space-y-2">
              <Label>Rule Type</Label>
              <Input :model-value="ruleTypeLabel(currentSharingRule.ruleType)" disabled />
            </div>

            <div class="space-y-2">
              <Label>Source Group</Label>
              <Input :model-value="groupName(currentSharingRule.sourceGroupId)" disabled />
            </div>

            <div class="space-y-2">
              <Label>Target Group</Label>
              <Select :model-value="state.targetGroupId" @update:model-value="onTargetGroupChange">
                <SelectTrigger>
                  <SelectValue placeholder="Select group" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="group in groups" :key="group.id" :value="group.id">
                    {{ group.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p v-if="errors.targetGroupId" class="text-sm text-destructive">{{ errors.targetGroupId }}</p>
            </div>

            <div class="space-y-2">
              <Label>Access Level</Label>
              <Select :model-value="state.accessLevel" @update:model-value="onAccessLevelChange">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="read">Read</SelectItem>
                  <SelectItem value="read_write">Read/Write</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card v-if="currentSharingRule.ruleType === 'criteria_based'">
          <CardContent class="pt-6 space-y-4">
            <h2 class="text-lg font-semibold">Criteria</h2>

            <div class="space-y-2">
              <Label for="criteriaField">Field</Label>
              <Input id="criteriaField" v-model="state.criteriaField" placeholder="status" />
              <p v-if="errors.criteriaField" class="text-sm text-destructive">{{ errors.criteriaField }}</p>
            </div>

            <div class="space-y-2">
              <Label for="criteriaOp">Operator</Label>
              <Input id="criteriaOp" v-model="state.criteriaOp" placeholder="=" />
              <p v-if="errors.criteriaOp" class="text-sm text-destructive">{{ errors.criteriaOp }}</p>
            </div>

            <div class="space-y-2">
              <Label for="criteriaValue">Value</Label>
              <Input id="criteriaValue" v-model="state.criteriaValue" placeholder="active" />
              <p v-if="errors.criteriaValue" class="text-sm text-destructive">{{ errors.criteriaValue }}</p>
            </div>
          </CardContent>
        </Card>

        <Separator />

        <div class="flex gap-2 items-center">
          <Button type="submit" :disabled="sharingRulesLoading">
            Save
          </Button>
          <IconButton
            :icon="X"
            tooltip="Cancel"
            variant="outline"
            @click="router.back()"
          />
        </div>
      </form>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete Rule?"
        description="This sharing rule will be permanently deleted."
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteRule"
      />
    </template>
  </div>
</template>
