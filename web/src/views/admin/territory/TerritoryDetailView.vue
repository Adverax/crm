<script setup lang="ts">
import { onMounted, watch, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { useTerritoryForm } from '@/composables/useTerritoryForm'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { storeToRefs } from 'pinia'

const props = defineProps<{
  territoryId: string
}>()

const router = useRouter()
const store = useTerritoryAdminStore()
const toast = useToast()
const {
  currentTerritory,
  territories,
  territoriesLoading,
  territoriesError,
  userAssignments,
  objectDefaults,
} = storeToRefs(store)
const { state, errors, validate, toUpdateRequest, initFrom } = useTerritoryForm()

const showDeleteDialog = ref(false)

async function loadData() {
  try {
    const territory = await store.fetchTerritory(props.territoryId)
    initFrom(territory)
    await Promise.all([
      store.fetchTerritories({ modelId: territory.modelId, perPage: 1000 }),
      store.fetchTerritoryUsers(props.territoryId),
      store.fetchObjectDefaults(props.territoryId),
    ])
  } catch (err) {
    toast.errorFromApi(err)
  }
}

onMounted(loadData)
watch(() => props.territoryId, loadData)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onParentChange(value: any) {
  state.parentId = value === '__none__' ? null : String(value)
}

async function onSave() {
  if (!validate()) return
  try {
    await store.updateTerritory(props.territoryId, toUpdateRequest())
    toast.success('Territory updated')
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onDelete() {
  try {
    await store.deleteTerritory(props.territoryId)
    toast.success('Territory deleted')
    router.push({ name: 'admin-territory-list' })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
  }
}

async function onRemoveUser(userId: string) {
  try {
    await store.unassignUser(props.territoryId, userId)
    toast.success('User removed from territory')
    await store.fetchTerritoryUsers(props.territoryId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

async function onRemoveObjectDefault(objectId: string) {
  try {
    await store.removeObjectDefault(props.territoryId, objectId)
    toast.success('Object default removed')
    await store.fetchObjectDefaults(props.territoryId)
  } catch (err) {
    toast.errorFromApi(err)
  }
}

const availableParents = computed(() =>
  territories.value.filter((t) => t.id !== props.territoryId),
)

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Territories', to: '/admin/territory/territories' },
  { label: currentTerritory.value?.label ?? '...' },
])
</script>

<template>
  <div>
    <div v-if="territoriesLoading && !currentTerritory" class="space-y-4">
      <Skeleton class="h-8 w-64" />
      <Skeleton class="h-64 w-full" />
    </div>

    <template v-else-if="currentTerritory">
      <PageHeader :title="currentTerritory.label" :breadcrumbs="breadcrumbs">
        <template #actions>
          <IconButton
            :icon="Trash2"
            tooltip="Delete territory"
            variant="destructive"
            @click="showDeleteDialog = true"
          />
        </template>
      </PageHeader>

      <ErrorAlert v-if="territoriesError" :message="territoriesError" class="mb-4" />

      <Tabs default-value="info" class="max-w-3xl">
        <TabsList>
          <TabsTrigger value="info">General</TabsTrigger>
          <TabsTrigger value="users">Users</TabsTrigger>
          <TabsTrigger value="objects">Objects</TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <form class="space-y-6 mt-4" @submit.prevent="onSave">
            <Card>
              <CardContent class="pt-6 space-y-4">
                <div class="space-y-2">
                  <Label>API Name</Label>
                  <Input :model-value="state.apiName" disabled />
                </div>

                <div class="space-y-2">
                  <Label for="label">Label</Label>
                  <Input id="label" v-model="state.label" />
                  <p v-if="errors.label" class="text-sm text-destructive">{{ errors.label }}</p>
                </div>

                <div class="space-y-2">
                  <Label for="parentId">Parent Territory</Label>
                  <Select :model-value="state.parentId ?? '__none__'" @update:model-value="onParentChange">
                    <SelectTrigger>
                      <SelectValue placeholder="No parent" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">No parent</SelectItem>
                      <SelectItem v-for="t in availableParents" :key="t.id" :value="t.id">
                        {{ t.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div class="space-y-2">
                  <Label for="description">Description</Label>
                  <Textarea id="description" v-model="state.description" rows="3" />
                </div>
              </CardContent>
            </Card>

            <Separator />

            <div class="flex gap-2 items-center">
              <Button type="submit" :disabled="territoriesLoading">
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

        <TabsContent value="users">
          <div class="mt-4 space-y-4">
            <Table v-if="userAssignments.length > 0">
              <TableHeader>
                <TableRow>
                  <TableHead>User ID</TableHead>
                  <TableHead>Assigned</TableHead>
                  <TableHead class="w-16" />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="assignment in userAssignments" :key="assignment.id">
                  <TableCell class="font-mono text-sm">{{ assignment.userId }}</TableCell>
                  <TableCell class="text-muted-foreground">
                    {{ new Date(assignment.createdAt).toLocaleDateString('en-US') }}
                  </TableCell>
                  <TableCell>
                    <IconButton
                      :icon="Trash2"
                      tooltip="Delete"
                      variant="ghost"
                      class="text-destructive"
                      @click="onRemoveUser(assignment.userId)"
                    />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <p v-else class="text-sm text-muted-foreground">
              No assigned users
            </p>
          </div>
        </TabsContent>

        <TabsContent value="objects">
          <div class="mt-4 space-y-4">
            <Table v-if="objectDefaults.length > 0">
              <TableHeader>
                <TableRow>
                  <TableHead>Object ID</TableHead>
                  <TableHead>Access Level</TableHead>
                  <TableHead class="w-16" />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="od in objectDefaults" :key="od.id">
                  <TableCell class="font-mono text-sm">{{ od.objectId }}</TableCell>
                  <TableCell>{{ od.accessLevel }}</TableCell>
                  <TableCell>
                    <IconButton
                      :icon="Trash2"
                      tooltip="Delete"
                      variant="ghost"
                      class="text-destructive"
                      @click="onRemoveObjectDefault(od.objectId)"
                    />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <p v-else class="text-sm text-muted-foreground">
              No object defaults for this territory
            </p>
          </div>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        :open="showDeleteDialog"
        title="Delete territory?"
        :description="`Territory '${currentTerritory.label}' (${currentTerritory.apiName}) will be permanently deleted.`"
        @update:open="showDeleteDialog = $event"
        @confirm="onDelete"
      />
    </template>
  </div>
</template>
