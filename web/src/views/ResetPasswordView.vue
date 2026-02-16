<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authApi } from '@/api/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const route = useRoute()
const router = useRouter()

const password = ref('')
const confirmPassword = ref('')
const error = ref<string | null>(null)
const success = ref(false)
const loading = ref(false)

const token = (route.query.token as string) || ''

async function onSubmit() {
  error.value = null

  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }

  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters'
    return
  }

  loading.value = true
  try {
    await authApi.resetPassword({ token, password: password.value })
    success.value = true
    setTimeout(() => router.push('/login'), 3000)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Password reset error'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-muted/30">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">New Password</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="success" class="space-y-4">
          <div class="rounded-md bg-green-50 p-3 text-sm text-green-800">
            Password changed successfully. Redirecting to login...
          </div>
        </div>

        <div v-else-if="!token" class="space-y-4">
          <div class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            The password reset link is invalid.
          </div>
          <RouterLink to="/login" class="block text-center text-sm text-primary hover:underline">
            Back to login
          </RouterLink>
        </div>

        <form v-else class="space-y-4" @submit.prevent="onSubmit">
          <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
          </div>

          <div class="space-y-2">
            <Label for="password">New Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="confirmPassword">Confirm Password</Label>
            <Input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>

          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Saving...' : 'Reset Password' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
