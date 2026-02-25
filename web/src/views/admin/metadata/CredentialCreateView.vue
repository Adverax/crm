<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { credentialsApi } from '@/api/credentials'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { CredentialType } from '@/types/credentials'

const router = useRouter()
const toast = useToast()
const submitting = ref(false)

const form = ref({
  code: '',
  name: '',
  description: '',
  type: 'api_key' as CredentialType,
  baseUrl: '',
  // api_key fields
  header: '',
  value: '',
  // basic fields
  username: '',
  password: '',
  // oauth2 fields
  clientId: '',
  clientSecret: '',
  tokenUrl: '',
  scope: '',
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onTypeChange(val: any) {
  form.value.type = String(val) as CredentialType
}

function buildAuthData() {
  switch (form.value.type) {
    case 'api_key':
      return { header: form.value.header, value: form.value.value }
    case 'basic':
      return { username: form.value.username, password: form.value.password }
    case 'oauth2_client':
      return {
        clientId: form.value.clientId,
        clientSecret: form.value.clientSecret,
        tokenUrl: form.value.tokenUrl,
        scope: form.value.scope || undefined,
      }
  }
}

async function onSubmit() {
  submitting.value = true
  try {
    const response = await credentialsApi.create({
      code: form.value.code,
      name: form.value.name,
      description: form.value.description || undefined,
      type: form.value.type,
      baseUrl: form.value.baseUrl,
      authData: buildAuthData(),
    })
    toast.success('Credential created')
    router.push({ name: 'admin-credential-detail', params: { credentialId: response.data.id } })
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  router.push({ name: 'admin-credentials' })
}

const breadcrumbs = computed(() => [
  { label: 'Admin', to: '/admin' },
  { label: 'Credentials', to: '/admin/metadata/credentials' },
  { label: 'Create' },
])
</script>

<template>
  <div>
    <PageHeader title="Create Credential" :breadcrumbs="breadcrumbs" />

    <form class="max-w-3xl space-y-6 mt-4" @submit.prevent="onSubmit">
      <Card>
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="code">Code</Label>
              <Input id="code" v-model="form.code" required placeholder="my_api" class="font-mono" data-testid="field-code" />
            </div>
            <div class="space-y-2">
              <Label>Type</Label>
              <Select :model-value="form.type" @update:model-value="onTypeChange">
                <SelectTrigger data-testid="field-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="api_key">API Key</SelectItem>
                  <SelectItem value="basic">Basic Auth</SelectItem>
                  <SelectItem value="oauth2_client">OAuth2 Client</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="space-y-2">
            <Label for="name">Name</Label>
            <Input id="name" v-model="form.name" required placeholder="My API" data-testid="field-name" />
          </div>
          <div class="space-y-2">
            <Label for="base-url">Base URL</Label>
            <Input id="base-url" v-model="form.baseUrl" required placeholder="https://api.example.com" data-testid="field-base-url" />
          </div>
          <div class="space-y-2">
            <Label for="description">Description</Label>
            <Textarea id="description" v-model="form.description" rows="2" data-testid="field-description" />
          </div>
        </CardContent>
      </Card>

      <!-- API Key Fields -->
      <Card v-if="form.type === 'api_key'">
        <CardContent class="pt-6 space-y-4">
          <Label class="text-base">API Key Authentication</Label>
          <div class="space-y-2">
            <Label for="header">Header Name</Label>
            <Input id="header" v-model="form.header" required placeholder="X-API-Key" data-testid="field-header" />
          </div>
          <div class="space-y-2">
            <Label for="api-value">API Key Value</Label>
            <Input id="api-value" v-model="form.value" required type="password" placeholder="sk-..." data-testid="field-value" />
          </div>
        </CardContent>
      </Card>

      <!-- Basic Auth Fields -->
      <Card v-if="form.type === 'basic'">
        <CardContent class="pt-6 space-y-4">
          <Label class="text-base">Basic Authentication</Label>
          <div class="space-y-2">
            <Label for="username">Username</Label>
            <Input id="username" v-model="form.username" required data-testid="field-username" />
          </div>
          <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input id="password" v-model="form.password" required type="password" data-testid="field-password" />
          </div>
        </CardContent>
      </Card>

      <!-- OAuth2 Fields -->
      <Card v-if="form.type === 'oauth2_client'">
        <CardContent class="pt-6 space-y-4">
          <Label class="text-base">OAuth2 Client Credentials</Label>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="client-id">Client ID</Label>
              <Input id="client-id" v-model="form.clientId" required data-testid="field-client-id" />
            </div>
            <div class="space-y-2">
              <Label for="client-secret">Client Secret</Label>
              <Input id="client-secret" v-model="form.clientSecret" required type="password" data-testid="field-client-secret" />
            </div>
          </div>
          <div class="space-y-2">
            <Label for="token-url">Token URL</Label>
            <Input id="token-url" v-model="form.tokenUrl" required placeholder="https://auth.example.com/oauth/token" data-testid="field-token-url" />
          </div>
          <div class="space-y-2">
            <Label for="scope">Scope (optional)</Label>
            <Input id="scope" v-model="form.scope" placeholder="read write" data-testid="field-scope" />
          </div>
        </CardContent>
      </Card>

      <div class="flex gap-2 items-center">
        <Button type="submit" :disabled="submitting" data-testid="submit-btn">
          Create
        </Button>
        <IconButton :icon="X" tooltip="Cancel" variant="outline" data-testid="cancel-btn" @click="onCancel" />
      </div>
    </form>
  </div>
</template>
