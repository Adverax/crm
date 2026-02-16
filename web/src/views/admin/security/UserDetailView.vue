<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { useUserForm } from '@/composables/useUserForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ActiveStatusBadge from '@/components/admin/security/ActiveStatusBadge.vue'
import UserPermissionSetsTab from '@/components/admin/security/UserPermissionSetsTab.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  userId: string
}>()

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { currentUser, profiles, roles, usersLoading, usersError } = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = useUserForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const [user] = await Promise.all([
      store.fetchUser(props.userId),
      store.fetchProfiles({ perPage: 1000 }),
      store.fetchRoles({ perPage: 1000 }),
    ])
    initFrom(user)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.userId, loadData)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onProfileChange(value: any) {
  state.profileId = String(value)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onRoleChange(value: any) {
  state.roleId = value === '__none__' ? null : String(value)
}

async function onSave() {
  if (!validate()) return
  try {
    await store.updateUser(props.userId, toUpdateRequest())
    toast.success('User updated')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDeleteUser() {
  try {
    await store.deleteUser(props.userId)
    toast.success('User deleted')
    router.push({ name: 'admin-users' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Users', to: '/admin/security/users' },
  { label: currentUser.value?.username ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="usersLoading && !currentUser" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentUser">
      <PageHeader :title="currentUser.username" :breadcrumbs="breadcrumbs">
        <template #actions>
          <ActiveStatusBadge :active="currentUser.isActive" />
          <IconButton
            :icon="Trash2"
            tooltip="Delete User"
            variant="destructive"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <ErrorAlert v-if="usersError" :message="usersError" class="mb-4" />

      <Tabs default-value="info">
        <TabsList>
          <TabsTrigger value="info">General</TabsTrigger>
          <TabsTrigger value="permission-sets">Permission Sets</TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <form class="max-w-2xl space-y-6 mt-4" @submit.prevent="onSave">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Credentials</h2>

                <div class="space-y-2">
                  <Label>Username</Label>
                  <Input :model-value="state.username" disabled />
                </div>

                <div class="space-y-2">
                  <Label for="email">Email</Label>
                  <Input id="email" v-model="state.email" type="email" />
                  <p v-if="errors.email" class="text-sm text-destructive">{{ errors.email }}</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Personal Information</h2>

                <div class="grid grid-cols-2 gap-4">
                  <div class="space-y-2">
                    <Label for="firstName">First Name</Label>
                    <Input id="firstName" v-model="state.firstName" />
                  </div>
                  <div class="space-y-2">
                    <Label for="lastName">Last Name</Label>
                    <Input id="lastName" v-model="state.lastName" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent class="pt-6 space-y-4">
                <h2 class="text-lg font-semibold">Security</h2>

                <div class="space-y-2">
                  <Label for="profileId">Profile</Label>
                  <Select :model-value="state.profileId || undefined" @update:model-value="onProfileChange">
                    <SelectTrigger>
                      <SelectValue placeholder="Select profile" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="profile in profiles" :key="profile.id" :value="profile.id">
                        {{ profile.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <p v-if="errors.profileId" class="text-sm text-destructive">{{ errors.profileId }}</p>
                </div>

                <div class="space-y-2">
                  <Label for="roleId">Role</Label>
                  <Select :model-value="state.roleId ?? '__none__'" @update:model-value="onRoleChange">
                    <SelectTrigger>
                      <SelectValue placeholder="No role" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">No role</SelectItem>
                      <SelectItem v-for="role in roles" :key="role.id" :value="role.id">
                        {{ role.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div class="flex items-center gap-3">
                  <Switch
                    :checked="state.isActive"
                    @update:checked="state.isActive = $event"
                  />
                  <Label>Active</Label>
                </div>
              </CardContent>
            </Card>

            <Separator />

            <div class="flex gap-2 items-center">
              <Button type="submit" :disabled="usersLoading">
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
        </TabsContent>

        <TabsContent value="permission-sets">
          <div class="mt-4">
            <UserPermissionSetsTab :user-id="props.userId" />
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete User?"
        :description="`User '${currentUser.username}' will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDeleteUser"
      />
    </template>
  </div>
</template>
