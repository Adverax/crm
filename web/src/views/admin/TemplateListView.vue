<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { templatesApi, type TemplateInfo } from '@/api/templates'

const toast = useToast()
const templates = ref<TemplateInfo[]>([])
const loading = ref(false)
const applying = ref<string | null>(null)

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Шаблоны' },
]

async function loadTemplates() {
  loading.value = true
  try {
    const response = await templatesApi.list()
    templates.value = response.data ?? []
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    loading.value = false
  }
}

async function applyTemplate(tmpl: TemplateInfo) {
  applying.value = tmpl.id
  try {
    await templatesApi.apply(tmpl.id)
    toast.success(`Шаблон "${tmpl.label}" применён`)
    await loadTemplates()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    applying.value = null
  }
}

onMounted(loadTemplates)
</script>

<template>
  <div>
    <PageHeader title="Шаблоны приложений" :breadcrumbs="breadcrumbs" />

    <p class="text-sm text-muted-foreground mb-6">
      Выберите шаблон для создания стандартных объектов и полей. Шаблон можно применить только один раз на пустую базу.
    </p>

    <div v-if="loading" class="space-y-4">
      <Skeleton v-for="i in 2" :key="i" class="h-40 w-full" />
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2">
      <Card v-for="tmpl in templates" :key="tmpl.id" class="flex flex-col">
        <CardHeader>
          <CardTitle>{{ tmpl.label }}</CardTitle>
          <CardDescription>{{ tmpl.description }}</CardDescription>
        </CardHeader>
        <CardContent class="flex-1 flex flex-col justify-between">
          <div class="text-sm text-muted-foreground mb-4">
            {{ tmpl.objects }} объектов, {{ tmpl.fields }} полей
          </div>
          <Button
            class="w-full"
            :disabled="applying !== null"
            @click="applyTemplate(tmpl)"
          >
            <template v-if="applying === tmpl.id">Применяется...</template>
            <template v-else>Применить</template>
          </Button>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
