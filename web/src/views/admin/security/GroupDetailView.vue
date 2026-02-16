<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
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
    personal: 'Personal',
    role: 'Role',
    role_and_subordinates: 'Role & Subordinates',
    public: 'Public',
    territory: 'Territory',
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
    toast.success('Group deleted')
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
    toast.success('Member added')
    addMemberUserId.value = ''
    await store.fetchGroupMembers(props.groupId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onRemoveMember(memberId: string) {
  try {
    await store.removeGroupMember(props.groupId, memberId)
    toast.success('Member removed')
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
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Groups', to: '/admin/security/groups' },
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
          <IconButton
            :icon="Trash2"
            tooltip="Delete Group"
            variant="destructive"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <ErrorAlert v-if="groupsError" :message="groupsError" class="mb-4" />

      <Tabs default-value="info">
        <TabsList>
          <TabsTrigger value="info">General</TabsTrigger>
          <TabsTrigger value="members">
            Members ({{ groupMembers.length }})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <Card class="max-w-2xl mt-4">
            <CardContent class="pt-6 space-y-4">
              <h2 class="text-lg font-semibold">General Information</h2>

              <div class="space-y-2">
                <Label>API Name</Label>
                <Input :model-value="currentGroup.apiName" disabled />
              </div>

              <div class="space-y-2">
                <Label>Label</Label>
                <Input :model-value="currentGroup.label" disabled />
              </div>

              <div class="space-y-2">
                <Label>Group Type</Label>
                <Input :model-value="groupTypeLabel(currentGroup.groupType)" disabled />
              </div>

              <div class="space-y-2">
                <Label>Created</Label>
                <Input :model-value="formatDate(currentGroup.createdAt)" disabled />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="members">
          <div class="mt-4 space-y-4">
            <Card class="max-w-2xl">
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Add Member</h2>
                <div class="flex gap-2 items-end">
                  <div class="flex-1 space-y-2">
                    <Label>User</Label>
                    <Select :model-value="addMemberUserId" @update:model-value="onUserSelect">
                      <SelectTrigger>
                        <SelectValue placeholder="Select user" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem v-for="user in users" :key="user.id" :value="user.id">
                          {{ user.firstName }} {{ user.lastName }} ({{ user.username }})
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <IconButton
                    :icon="Plus"
                    tooltip="Add"
                    variant="default"
                    :disabled="!addMemberUserId || groupsLoading"
                    @click="onAddMember"
                  />
                </div>
              </CardContent>
            </Card>

            <EmptyState
              v-if="groupMembers.length === 0"
              title="No Members"
              description="Add users to this group"
            />

            <Table v-else>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Added</TableHead>
                  <TableHead class="w-16" />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="member in groupMembers" :key="member.id">
                  <TableCell class="font-medium">{{ getUserName(member.memberUserId) }}</TableCell>
                  <TableCell class="text-muted-foreground">{{ formatDate(member.createdAt) }}</TableCell>
                  <TableCell>
                    <IconButton
                      :icon="Trash2"
                      tooltip="Delete"
                      variant="ghost"
                      class="text-destructive hover:text-destructive"
                      @click="onRemoveMember(member.id)"
                    />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete Group?"
        :description="`Group '${currentGroup.label}' (${currentGroup.apiName}) will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteGroup"
      />
    </template>
  </div>
</template>
