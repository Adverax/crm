<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, ChevronLeft, ChevronRight, MoreVertical } from 'lucide-vue-next'
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
import { Skeleton } from '@/components/ui/skeleton'
import type { Profile } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { profiles, profilesPagination, profilesLoading } = storeToRefs(store)

const deleteTarget = ref<Profile | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  profilesPagination,
  (page) => loadProfiles(page),
)

function loadProfiles(page = 1) {
  store.fetchProfiles({ page, perPage: 20 }).catch((err) => toast.errorFromApi(err))
}

onMounted(() => loadProfiles())

function goToDetail(profile: Profile) {
  router.push({ name: 'admin-profile-detail', params: { profileId: profile.id } })
}

function confirmDelete(profile: Profile) {
  deleteTarget.value = profile
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteProfile(deleteTarget.value.id)
    toast.success('Profile deleted')
    loadProfiles()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Profiles' },
]
</script>

<template>
  <div>
    <PageHeader title="Profiles" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create Profile"
          variant="default"
          @click="router.push({ name: 'admin-profile-create' })"
        />
      </template>
    </PageHeader>

    <div v-if="profilesLoading && profiles.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!profilesLoading && profiles.length === 0"
      title="No Profiles"
      description="Create your first security profile"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create Profile"
          variant="default"
          @click="router.push({ name: 'admin-profile-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Label</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="profile in profiles"
            :key="profile.id"
            class="cursor-pointer"
            @click="goToDetail(profile)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-profile-detail', params: { profileId: profile.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ profile.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ profile.label }}</TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(profile.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(profile)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(profile)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="profilesPagination && profilesPagination.totalPages > 1" class="flex items-center justify-between mt-4">
        <span class="text-sm text-muted-foreground">{{ pageInfo }}</span>
        <div class="flex gap-2">
          <IconButton
            :icon="ChevronLeft"
            tooltip="Back"
            variant="outline"
            :disabled="isFirstPage"
            @click="prevPage"
          />
          <IconButton
            :icon="ChevronRight"
            tooltip="Next"
            variant="outline"
            :disabled="isLastPage"
            @click="nextPage"
          />
        </div>
      </div>
    </template>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete Profile?"
      :description="`Profile '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
