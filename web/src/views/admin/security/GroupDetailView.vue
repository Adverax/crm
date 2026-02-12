<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import EmptyState from '@/components/admin/EmptyState.vue'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  groupId: string
}>()

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { currentGroup, groupsLoading, groupsError, groupMembers, users } = storeToRefs(store)

const showDeleteDialog = ref(false)
const addMemberUserId = ref('')

function groupTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    personal: 'Персональная',
    role: 'Роль',
    role_and_subordinates: 'Роль и подчинённые',
    public: 'Публичная',
    territory: 'Территория',
  }
  return labels[type] ?? type
}

async function loadData() {
  try {
    await store.fetchGroup(props.groupId)
    await store.fetchGroupMembers(props.groupId)
    await store.fetchUsers({ perPage: 1000 })
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.groupId, loadData)

async function onDeleteGroup() {
  try {
    await store.deleteGroup(props.groupId)
    toast.success('Группа удалена')
    router.push({ name: 'admin-groups' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onUserSelect(value: any) {
  addMemberUserId.value = String(value)
}

async function onAddMember() {
  if (!addMemberUserId.value) return
  try {
    await store.addGroupMember(props.groupId, { memberUserId: addMemberUserId.value })
    toast.success('Участник добавлен')
    addMemberUserId.value = ''
    await store.fetchGroupMembers(props.groupId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onRemoveMember(memberId: string) {
  try {
    await store.removeGroupMember(props.groupId, memberId)
    toast.success('Участник удалён')
    await store.fetchGroupMembers(props.groupId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

function getUserName(userId: string | null): string {
  if (!userId) return '—'
  const user = users.value.find((u) => u.id === userId)
  return user ? `${user.firstName} ${user.lastName} (${user.username})` : userId
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = computed(() => [
  { label: 'Админ', to: '/admin' },
  { label: 'Группы', to: '/admin/security/groups' },
  { label: currentGroup.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="groupsLoading && !currentGroup" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentGroup">
      <PageHeader :title="currentGroup.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <Button
            variant="destructive"
            size="sm"
            @click="showDeleteDialog = true"
          >
            Удалить группу
          </Button>
        </template>
      </PageHeader>

      <ErrorAlert v-if="groupsError" :message="groupsError" class="mb-4" />

      <Tabs default-value="info">
        <TabsList>
          <TabsTrigger value="info">Основное</TabsTrigger>
          <TabsTrigger value="members">
            Участники ({{ groupMembers.length }})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <Card class="max-w-2xl mt-4">
            <CardContent class="pt-6 space-y-4">
              <h2 class="text-lg font-semibold">Основная информация</h2>

              <div class="space-y-2">
                <Label>API Name</Label>
                <Input :model-value="currentGroup.apiName" disabled />
              </div>

              <div class="space-y-2">
                <Label>Название</Label>
                <Input :model-value="currentGroup.label" disabled />
              </div>

              <div class="space-y-2">
                <Label>Тип группы</Label>
                <Input :model-value="groupTypeLabel(currentGroup.groupType)" disabled />
              </div>

              <div class="space-y-2">
                <Label>Создана</Label>
                <Input :model-value="formatDate(currentGroup.createdAt)" disabled />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="members">
          <div class="mt-4 space-y-4">
            <Card class="max-w-2xl">
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Добавить участника</h2>
                <div class="flex gap-2 items-end">
                  <div class="flex-1 space-y-2">
                    <Label>Пользователь</Label>
                    <Select :model-value="addMemberUserId" @update:model-value="onUserSelect">
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите пользователя" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem v-for="user in users" :key="user.id" :value="user.id">
                          {{ user.firstName }} {{ user.lastName }} ({{ user.username }})
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button :disabled="!addMemberUserId || groupsLoading" @click="onAddMember">
                    Добавить
                  </Button>
                </div>
              </CardContent>
            </Card>

            <EmptyState
              v-if="groupMembers.length === 0"
              title="Нет участников"
              description="Добавьте пользователей в группу"
            />

            <Table v-else>
              <TableHeader>
                <TableRow>
                  <TableHead>Пользователь</TableHead>
                  <TableHead>Добавлен</TableHead>
                  <TableHead class="w-16" />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="member in groupMembers" :key="member.id">
                  <TableCell class="font-medium">{{ getUserName(member.memberUserId) }}</TableCell>
                  <TableCell class="text-muted-foreground">{{ formatDate(member.createdAt) }}</TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="text-destructive"
                      @click="onRemoveMember(member.id)"
                    >
                      Удалить
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Удалить группу?"
        :description="`Группа «${currentGroup.label}» (${currentGroup.apiName}) будет удалена без возможности восстановления.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteGroup"
      />
    </template>
  </div>
</template>
