<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useMetadataStore } from '@/stores/metadata'
import { useSharingRuleForm } from '@/composables/useSharingRuleForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
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
import { storeToRefs } from 'pinia'
import type { RuleType, AccessLevel } from '@/types/security'

const router = useRouter()
const securityStore = useSecurityAdminStore()
const metadataStore = useMetadataStore()
const toast = useToast()

const { sharingRulesLoading, sharingRulesError, groups } = storeToRefs(securityStore)
const { objects } = storeToRefs(metadataStore)
const { state, errors, validate, toCreateRequest } = useSharingRuleForm()

onMounted(async () => {
  try {
    await metadataStore.fetchObjects({ perPage: 1000 })
    await securityStore.fetchGroups({ perPage: 1000 })
  } catch (err) {
    toast.errorFromApi(err)
  }
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onObjectChange(value: any) {
  state.objectId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onRuleTypeChange(value: any) {
  state.ruleType = String(value) as RuleType
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onSourceGroupChange(value: any) {
  state.sourceGroupId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onTargetGroupChange(value: any) {
  state.targetGroupId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onAccessLevelChange(value: any) {
  state.accessLevel = String(value) as AccessLevel
}

async function onSubmit() {
  if (!validate()) return
  try {
    const created = await securityStore.createSharingRule(toCreateRequest())
    toast.success('Rule created')
    router.push({ name: 'admin-sharing-rule-detail', params: { ruleId: created.id } })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Sharing Rules', to: '/admin/security/sharing-rules' },
  { label: 'New Rule' },
]
</script>

<template>
  <div>
    <PageHeader title="Create Sharing Rule" :breadcrumbs="breadcrumbs" />

    <ErrorAlert v-if="sharingRulesError" :message="sharingRulesError" class="mb-4" />

    <form class="max-w-2xl space-y-6" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <h2 class="text-lg font-semibold">General Information</h2>

          <div class="space-y-2">
            <Label>Object</Label>
            <Select :model-value="state.objectId" @update:model-value="onObjectChange">
              <SelectTrigger>
                <SelectValue placeholder="Select object" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="obj in objects" :key="obj.id" :value="obj.id">
                  {{ obj.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="errors.objectId" class="text-sm text-destructive">{{ errors.objectId }}</p>
          </div>

          <div class="space-y-2">
            <Label>Rule Type</Label>
            <Select :model-value="state.ruleType" @update:model-value="onRuleTypeChange">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="owner_based">Owner-based</SelectItem>
                <SelectItem value="criteria_based">Criteria-based</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label>Source Group</Label>
            <Select :model-value="state.sourceGroupId" @update:model-value="onSourceGroupChange">
              <SelectTrigger>
                <SelectValue placeholder="Select group" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="group in groups" :key="group.id" :value="group.id">
                  {{ group.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="errors.sourceGroupId" class="text-sm text-destructive">{{ errors.sourceGroupId }}</p>
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

      <Card v-if="state.ruleType === 'criteria_based'">
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
          Create
        </Button>
        <IconButton
          :icon="X"
          tooltip="Cancel"
          variant="outline"
          @click="router.back()"
        />
      </div>
    </form>
  </div>
</template>
