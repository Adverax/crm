<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Input } from '@/components/ui/input'
import { useFunctionsStore } from '@/stores/functions'

const emit = defineEmits<{
  insert: [text: string]
}>()

const functionsStore = useFunctionsStore()
onMounted(() => functionsStore.ensureLoaded())

const search = ref('')

interface PickerItem {
  label: string
  value: string
  detail?: string
}

interface PickerGroup {
  label: string
  items: PickerItem[]
}

const builtinGroups: PickerGroup[] = [
  {
    label: 'String',
    items: [
      { label: 'size()', value: 'size()', detail: 'Length' },
      { label: 'contains()', value: 'contains()', detail: 'Contains' },
      { label: 'startsWith()', value: 'startsWith()', detail: 'Starts with' },
      { label: 'endsWith()', value: 'endsWith()', detail: 'Ends with' },
      { label: 'matches()', value: 'matches()', detail: 'Regular expression' },
    ],
  },
  {
    label: 'Type Casting',
    items: [
      { label: 'int()', value: 'int()', detail: 'To integer' },
      { label: 'double()', value: 'double()', detail: 'To double' },
      { label: 'string()', value: 'string()', detail: 'To string' },
      { label: 'bool()', value: 'bool()', detail: 'To boolean' },
    ],
  },
  {
    label: 'Time',
    items: [
      { label: 'duration()', value: 'duration()', detail: 'Duration' },
      { label: 'timestamp()', value: 'timestamp()', detail: 'Timestamp' },
    ],
  },
  {
    label: 'General',
    items: [
      { label: 'has()', value: 'has()', detail: 'Field presence' },
      { label: 'type()', value: 'type()', detail: 'Value type' },
    ],
  },
]

const groups = computed<PickerGroup[]>(() => {
  const q = search.value.toLowerCase()
  const result: PickerGroup[] = []

  // Custom functions (fn.*)
  const customFns = functionsStore.functions
  if (customFns.length > 0) {
    const items = customFns
      .map((fn) => {
        const params = fn.params ?? []
        const paramStr = params.map((p) => p.name).join(', ')
        return {
          label: `fn.${fn.name}(${paramStr})`,
          value: `fn.${fn.name}(${paramStr})`,
          detail: fn.description ?? undefined,
        }
      })
      .filter(
        (i) =>
          !q ||
          i.label.toLowerCase().includes(q) ||
          (i.detail?.toLowerCase().includes(q) ?? false),
      )
    if (items.length > 0) {
      result.push({ label: 'Custom (fn.*)', items })
    }
  }

  // Built-in groups
  for (const group of builtinGroups) {
    const items = group.items.filter(
      (i) =>
        !q ||
        i.label.toLowerCase().includes(q) ||
        (i.detail?.toLowerCase().includes(q) ?? false),
    )
    if (items.length > 0) {
      result.push({ label: group.label, items })
    }
  }

  return result
})

function onInsert(value: string) {
  emit('insert', value)
}
</script>

<template>
  <div class="space-y-3" data-testid="function-picker">
    <Input
      v-model="search"
      placeholder="Search function..."
      class="h-8 text-xs"
      data-testid="function-picker-search"
    />

    <div
      class="max-h-64 overflow-y-auto space-y-2"
    >
      <div v-for="group in groups" :key="group.label" class="space-y-1">
        <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          {{ group.label }}
        </div>
        <button
          v-for="item in group.items"
          :key="item.value"
          type="button"
          class="w-full text-left px-2 py-1 text-xs rounded hover:bg-accent hover:text-accent-foreground transition-colors"
          :title="item.detail"
          @click="onInsert(item.value)"
        >
          <code class="text-xs">{{ item.label }}</code>
        </button>
      </div>
    </div>

    <div v-if="groups.length === 0" class="text-xs text-muted-foreground">
      No functions found
    </div>
  </div>
</template>
