<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useToast } from '@/composables/useToast'
import PsTypeBadge from '@/components/admin/security/PsTypeBadge.vue'
import AssignPermissionSetDialog from '@/components/admin/security/AssignPermissionSetDialog.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import type { PermissionSetAssignment } from '@/types/security'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  userId: string
}>()

const store = useSecurityAdminStore()
const toast = useToast()
const { userPermissionSets, permissionSets, usersLoading } = storeToRefs(store)

const showAssignDialog = ref(false)
const revokeTarget = ref<PermissionSetAssignment | null>(null)
const showRevokeDialog = ref(false)

onMounted(async () => {
  try {
    await Promise.all([
      store.fetchUserPermissionSets(props.userId),
      store.fetchPermissionSets({ perPage: 1000 }),
    ])
  } catch (err) {
    toast.errorFromApi(err)
  }
})

const enrichedAssignments = computed(() =>
  userPermissionSets.value.map((a) => {
    const ps = permissionSets.value.find((p) => p.id === a.permissionSetId)
    return { ...a, psLabel: ps?.label ?? a.permissionSetId, psType: ps?.psType ?? 'grant' as const }
  }),
)

function confirmRevoke(assignment: PermissionSetAssignment) {
  revokeTarget.value = assignment
  showRevokeDialog.value = true
}

async function onRevokeConfirmed() {
  if (!revokeTarget.value) return
  try {
    await store.revokePermissionSet(props.userId, revokeTarget.value.id)
    toast.success('Permission set revoked')
    await store.fetchUserPermissionSets(props.userId)
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showRevokeDialog.value = false
    revokeTarget.value = null
  }
}

async function onAssigned() {
  showAssignDialog.value = false
  await store.fetchUserPermissionSets(props.userId)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

function getRevokeLabel(): string {
  if (!revokeTarget.value) return ''
  const ps = permissionSets.value.find((p) => p.id === revokeTarget.value!.permissionSetId)
  return ps?.label ?? revokeTarget.value.permissionSetId
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold">Permission Sets</h2>
      <Button size="sm" @click="showAssignDialog = true">
        Assign Set
      </Button>
    </div>

    <div v-if="usersLoading && userPermissionSets.length === 0" class="space-y-3">
      <Skeleton v-for="i in 3" :key="i" class="h-12 w-full" />
    </div>

    <div v-else-if="userPermissionSets.length === 0" class="text-sm text-muted-foreground py-8 text-center">
      No additional permission sets assigned to this user
    </div>

    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Assigned</TableHead>
          <TableHead class="w-24" />
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="assignment in enrichedAssignments" :key="assignment.id">
          <TableCell class="font-medium">{{ assignment.psLabel }}</TableCell>
          <TableCell>
            <PsTypeBadge :type="assignment.psType" />
          </TableCell>
          <TableCell class="text-muted-foreground">{{ formatDate(assignment.createdAt) }}</TableCell>
          <TableCell>
            <Button
              variant="ghost"
              size="sm"
              class="text-destructive"
              @click="confirmRevoke(assignment)"
            >
              Revoke
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <AssignPermissionSetDialog
      :open="showAssignDialog"
      :user-id="props.userId"
      :assigned-sets="userPermissionSets"
      @update:open="showAssignDialog = $event"
      @assigned="onAssigned"
    />

    <ConfirmDialog
      :open="showRevokeDialog"
      title="Revoke permission set?"
      :description="`Permission set '${getRevokeLabel()}' will be revoked from the user.`"
      confirm-label="Revoke"
      @update:open="showRevokeDialog = $event"
      @confirm="onRevokeConfirmed"
    />
  </div>
</template>
