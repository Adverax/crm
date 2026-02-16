<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useMetadataStore } from '@/stores/metadata'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, MoreVertical } from 'lucide-vue-next'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import type { SharingRule } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const securityStore = useSecurityAdminStore()
const metadataStore = useMetadataStore()
const toast = useToast()

const { sharingRules, sharingRulesLoading, groups } = storeToRefs(securityStore)
const { objects } = storeToRefs(metadataStore)

const selectedObjectId = ref('')
const deleteTarget = ref<SharingRule | null>(null)
const showDeleteDialog = ref(false)

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
  selectedObjectId.value = String(value)
  loadRules()
}

function loadRules() {
  if (!selectedObjectId.value) return
  securityStore.fetchSharingRules({ objectId: selectedObjectId.value }).catch((err) => toast.errorFromApi(err))
}

function goToDetail(rule: SharingRule) {
  router.push({ name: 'admin-sharing-rule-detail', params: { ruleId: rule.id } })
}

function confirmDelete(rule: SharingRule) {
  deleteTarget.value = rule
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await securityStore.deleteSharingRule(deleteTarget.value.id)
    toast.success('Rule deleted')
    loadRules()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function ruleTypeLabel(type: string): string {
  return type === 'owner_based' ? 'Owner-based' : 'Criteria-based'
}

function accessLevelLabel(level: string): string {
  return level === 'read' ? 'Read' : 'Read/Write'
}

function groupName(groupId: string): string {
  const group = groups.value.find((g) => g.id === groupId)
  return group?.label ?? groupId
}

function objectName(objectId: string): string {
  const obj = objects.value.find((o) => o.id === objectId)
  return obj?.label ?? objectId
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Sharing Rules' },
]
</script>

<template>
  <div>
    <PageHeader title="Sharing Rules" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create Rule"
          variant="default"
          @click="router.push({ name: 'admin-sharing-rule-create' })"
        />
      </template>
    </PageHeader>

    <div class="max-w-xs mb-6 space-y-2">
      <Label>Object</Label>
      <Select :model-value="selectedObjectId" @update:model-value="onObjectChange">
        <SelectTrigger>
          <SelectValue placeholder="Select object" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="obj in objects" :key="obj.id" :value="obj.id">
            {{ obj.label }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="!selectedObjectId" class="text-sm text-muted-foreground">
      Select an object to view sharing rules.
    </div>

    <div v-else-if="sharingRulesLoading && sharingRules.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!sharingRulesLoading && sharingRules.length === 0"
      title="No Rules"
      :description="`No sharing rules for object '${objectName(selectedObjectId)}'`"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create Rule"
          variant="default"
          @click="router.push({ name: 'admin-sharing-rule-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Type</TableHead>
            <TableHead>Source Group</TableHead>
            <TableHead>Target Group</TableHead>
            <TableHead>Access</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="rule in sharingRules"
            :key="rule.id"
            class="cursor-pointer"
            @click="goToDetail(rule)"
          >
            <TableCell>{{ ruleTypeLabel(rule.ruleType) }}</TableCell>
            <TableCell class="font-medium">{{ groupName(rule.sourceGroupId) }}</TableCell>
            <TableCell class="font-medium">{{ groupName(rule.targetGroupId) }}</TableCell>
            <TableCell>{{ accessLevelLabel(rule.accessLevel) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(rule)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(rule)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </template>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete Rule?"
      description="This sharing rule will be permanently deleted."
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
