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
    toast.success('Правило удалено')
    loadRules()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function ruleTypeLabel(type: string): string {
  return type === 'owner_based' ? 'По владельцу' : 'По критерию'
}

function accessLevelLabel(level: string): string {
  return level === 'read' ? 'Чтение' : 'Чтение/Запись'
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
  { label: 'Админ', to: '/admin' },
  { label: 'Правила совместного доступа' },
]
</script>

<template>
  <div>
    <PageHeader title="Правила совместного доступа" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-sharing-rule-create' })">
          Создать правило
        </Button>
      </template>
    </PageHeader>

    <div class="max-w-xs mb-6 space-y-2">
      <Label>Объект</Label>
      <Select :model-value="selectedObjectId" @update:model-value="onObjectChange">
        <SelectTrigger>
          <SelectValue placeholder="Выберите объект" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="obj in objects" :key="obj.id" :value="obj.id">
            {{ obj.label }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="!selectedObjectId" class="text-sm text-muted-foreground">
      Выберите объект для просмотра правил совместного доступа.
    </div>

    <div v-else-if="sharingRulesLoading && sharingRules.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!sharingRulesLoading && sharingRules.length === 0"
      title="Нет правил"
      :description="`Для объекта «${objectName(selectedObjectId)}» нет правил совместного доступа`"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-sharing-rule-create' })">
          Создать правило
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Тип</TableHead>
            <TableHead>Группа-источник</TableHead>
            <TableHead>Группа-получатель</TableHead>
            <TableHead>Доступ</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(rule)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(rule)"
                  >
                    Удалить
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
      title="Удалить правило?"
      description="Правило совместного доступа будет удалено без возможности восстановления."
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
