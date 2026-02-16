<script setup lang="ts">
import { ref } from 'vue'
import { authApi } from '@/api/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const email = ref('')
const error = ref<string | null>(null)
const success = ref(false)
const loading = ref(false)

async function onSubmit() {
  error.value = null
  loading.value = true
  try {
    await authApi.forgotPassword({ email: email.value })
    success.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to send'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-muted/30">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Reset Password</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="success" class="space-y-4">
          <div class="rounded-md bg-green-50 p-3 text-sm text-green-800">
            If this email is registered, a password reset link has been sent.
          </div>
          <RouterLink to="/login" class="block text-center text-sm text-primary hover:underline">
            Back to login
          </RouterLink>
        </div>

        <form v-else class="space-y-4" @submit.prevent="onSubmit">
          <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
          </div>

          <p class="text-sm text-muted-foreground">
            Enter the email associated with your account.
          </p>

          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              autocomplete="email"
              required
            />
          </div>

          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Sending...' : 'Submit' }}
          </Button>

          <div class="text-center text-sm">
            <RouterLink to="/login" class="text-primary hover:underline">
              Back to login
            </RouterLink>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
