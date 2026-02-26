<script setup lang="ts">
import { ref, computed } from 'vue'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface ObjectInfo {
  apiName: string
  label: string
}

interface FieldInfo {
  apiName: string
  label: string
  fieldType: string
}

const props = withDefaults(
  defineProps<{
    objects?: ObjectInfo[]
    fields?: FieldInfo[]
    currentObject?: string | null
  }>(),
  {
    objects: () => [],
    fields: () => [],
    currentObject: null,
  },
)

const emit = defineEmits<{
  insert: [text: string]
  'select-object': [objectName: string]
}>()

const search = ref('')
const activeTab = ref('objects')

const filteredObjects = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return props.objects
  return props.objects.filter(
    (o) => o.apiName.toLowerCase().includes(q) || o.label.toLowerCase().includes(q),
  )
})

const filteredFields = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return props.fields
  return props.fields.filter(
    (f) => f.apiName.toLowerCase().includes(q) || f.label.toLowerCase().includes(q),
  )
})

function onSelectObject(name: string) {
  emit('select-object', name)
  search.value = ''
  activeTab.value = 'fields'
}

function onInsertField(name: string) {
  emit('insert', name)
}
</script>

<template>
  <div class="space-y-2" data-testid="soql-object-picker">
    <Input
      v-model="search"
      placeholder="Search..."
      class="h-8 text-xs"
    />

    <Tabs v-model="activeTab" class="w-full">
      <TabsList class="h-8 w-full">
        <TabsTrigger value="objects" class="text-xs h-7 flex-1">
          Objects
        </TabsTrigger>
        <TabsTrigger value="fields" class="text-xs h-7 flex-1">
          Fields
        </TabsTrigger>
      </TabsList>

      <TabsContent value="objects" class="mt-2 max-h-60 overflow-y-auto">
        <div v-if="filteredObjects.length > 0" class="space-y-0.5">
          <button
            v-for="obj in filteredObjects"
            :key="obj.apiName"
            type="button"
            class="w-full text-left px-2 py-1 text-xs rounded hover:bg-accent hover:text-accent-foreground transition-colors"
            :class="{ 'bg-accent/50': obj.apiName === currentObject }"
            :title="obj.label"
            @click="onSelectObject(obj.apiName)"
          >
            <code class="text-xs">{{ obj.apiName }}</code>
            <span class="text-muted-foreground ml-1">{{ obj.label }}</span>
          </button>
        </div>
        <div v-else class="text-xs text-muted-foreground px-2 py-1">
          No objects available
        </div>
      </TabsContent>

      <TabsContent value="fields" class="mt-2">
        <div
          v-if="currentObject"
          class="text-xs text-muted-foreground px-2 pb-1 border-b mb-1"
        >
          {{ currentObject }}
        </div>
        <div class="max-h-56 overflow-y-auto">
          <div v-if="filteredFields.length > 0" class="space-y-0.5">
            <button
              v-for="field in filteredFields"
              :key="field.apiName"
              type="button"
              class="w-full text-left px-2 py-1 text-xs rounded hover:bg-accent hover:text-accent-foreground transition-colors"
              :title="field.label"
              @click="onInsertField(field.apiName)"
            >
              <code class="text-xs">{{ field.apiName }}</code>
              <span class="text-muted-foreground ml-1">{{ field.fieldType }}</span>
            </button>
          </div>
          <div v-else class="text-xs text-muted-foreground px-2 py-1">
            {{ fields.length === 0 ? 'Click an object to see its fields' : 'No matching fields' }}
          </div>
        </div>
      </TabsContent>
    </Tabs>
  </div>
</template>
