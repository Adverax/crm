<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useToast } from '@/composables/useToast'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { storeToRefs } from 'pinia'
import type { PermissionSetAssignment } from '@/types/security'

const props = defineProps<{
  open: boolean
  userId: string
  assignedSets: PermissionSetAssignment[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  assigned: []
}>()

const store = useSecurityAdminStore()
const toast = useToast()
const { permissionSets, permissionSetsLoading } = storeToRefs(store)

const selectedPsId = ref<string>('')

const assignedIds = computed(() =>
  new Set(props.assignedSets.map((a) => a.permissionSetId)),
)

const availableSets = computed(() =>
  permissionSets.value.filter((ps) => !assignedIds.value.has(ps.id)),
)

onMounted(() => {
  store.fetchPermissionSets({ perPage: 1000 }).catch((err) => toast.errorFromApi(err))
})

function onCancel() {
  selectedPsId.value = ''
  emit('update:open', false)
}

async function onAssign() {
  if (!selectedPsId.value) return
  try {
    await store.assignPermissionSet(props.userId, selectedPsId.value)
    toast.success('Permission set assigned')
    selectedPsId.value = ''
    emit('assigned')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onSelect(value: any) {
  selectedPsId.value = String(value)
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Assign Permission Set</DialogTitle>
        <DialogDescription>
          Select a permission set to assign to the user
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-4">
        <div class="space-y-2">
          <Label>Permission Set</Label>
          <Select :model-value="selectedPsId || undefined" @update:model-value="onSelect">
            <SelectTrigger>
              <SelectValue placeholder="Select a set" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="ps in availableSets" :key="ps.id" :value="ps.id">
                {{ ps.label }} ({{ ps.psType }})
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <p v-if="availableSets.length === 0" class="text-sm text-muted-foreground">
          All available sets are already assigned
        </p>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="permissionSetsLoading" @click="onCancel">
          Cancel
        </Button>
        <Button :disabled="!selectedPsId || permissionSetsLoading" @click="onAssign">
          Assign
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
