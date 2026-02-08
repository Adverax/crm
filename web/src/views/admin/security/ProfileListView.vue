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
const { profiles, profilesPagination, isLoading } = storeToRefs(store)

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
    toast.success('Профиль удалён')
    loadProfiles()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Профили' },
]
</script>

<template>
  <div>
    <PageHeader title="Профили" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-profile-create' })">
          Создать профиль
        </Button>
      </template>
    </PageHeader>

    <div v-if="isLoading && profiles.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!isLoading && profiles.length === 0"
      title="Нет профилей"
      description="Создайте первый профиль безопасности"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-profile-create' })">
          Создать профиль
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Название</TableHead>
            <TableHead>Создан</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(profile)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(profile)"
                  >
                    Удалить
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
          <Button variant="outline" size="sm" :disabled="isFirstPage" @click="prevPage">
            Назад
          </Button>
          <Button variant="outline" size="sm" :disabled="isLastPage" @click="nextPage">
            Вперёд
          </Button>
        </div>
      </div>
    </template>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Удалить профиль?"
      :description="`Профиль «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалён без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
