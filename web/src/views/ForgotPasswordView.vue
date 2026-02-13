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
    error.value = err instanceof Error ? err.message : 'Ошибка отправки'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-muted/30">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Сброс пароля</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="success" class="space-y-4">
          <div class="rounded-md bg-green-50 p-3 text-sm text-green-800">
            Если указанный email зарегистрирован, вам отправлена ссылка для сброса пароля.
          </div>
          <RouterLink to="/login" class="block text-center text-sm text-primary hover:underline">
            Вернуться ко входу
          </RouterLink>
        </div>

        <form v-else class="space-y-4" @submit.prevent="onSubmit">
          <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
          </div>

          <p class="text-sm text-muted-foreground">
            Введите email, привязанный к вашему аккаунту.
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
            {{ loading ? 'Отправка...' : 'Отправить' }}
          </Button>

          <div class="text-center text-sm">
            <RouterLink to="/login" class="text-primary hover:underline">
              Вернуться ко входу
            </RouterLink>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
