<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { proceduresApi } from '@/api/procedures'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const form = ref({
  code: '',
  name: '',
  description: '',
})

async function onSubmit() {
  submitting.value = true
  try {
    const response = await proceduresApi.create({
      code: form.value.code,
      name: form.value.name,
      description: form.value.description || undefined,
    })
    toast.success('Procedure created')
    router.push({ name: 'admin-procedure-detail', params: { procedureId: response.data.procedure.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-procedures' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Procedures', to: '/admin/metadata/procedures' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Procedure" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="space-y-2">
            <Label for="code">Code</Label>
            <Input
              id="code"
              v-model="form.code"
              required
              placeholder="my_procedure"
              class="font-mono"
              data-testid="field-code"
            />
            <p class="text-xs text-muted-foreground">
              Lowercase letters, numbers, and underscores. Must start with a letter.
            </p>
          </div>

          <div class="space-y-2">
            <Label for="name">Name</Label>
            <Input
              id="name"
              v-model="form.name"
              required
              placeholder="My Procedure"
              data-testid="field-name"
            />
          </div>

          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea
              id="description"
              v-model="form.description"
              rows="2"
              data-testid="field-description"
            />
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton
          :icon="X"
          tooltip="Cancel"
          variant="outline"
          data-testid="cancel-btn"
          @click="onCancel"
        />
      </div>
    </form>
  </div>
</template>
