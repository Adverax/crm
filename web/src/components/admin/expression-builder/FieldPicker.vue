<script setup lang="ts">
import { ref, computed } from 'vue'
import { Input } from '@/components/ui/input'
import type { CelContext, FunctionParam } from '@/types/functions'

interface DescribeField {
  apiName: string
  label: string
  fieldType: string
}

const props = withDefaults(
  defineProps<{
    fields?: DescribeField[]
    params?: FunctionParam[]
    context: CelContext
  }>(),
  {
    fields: () => [],
    params: () => [],
  },
)

const emit = defineEmits<{
  insert: [text: string]
}>()

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

const groups = computed<PickerGroup[]>(() => {
  const result: PickerGroup[] = []
  const q = search.value.toLowerCase()

  // Record fields
  if (props.fields.length > 0 && props.context !== 'function_body') {
    const items = props.fields
      .map((f) => ({
        label: `record.${f.apiName}`,
        value: `record.${f.apiName}`,
        detail: f.label,
      }))
      .filter((i) => !q || i.label.toLowerCase().includes(q) || (i.detail?.toLowerCase().includes(q) ?? false))
    if (items.length > 0) {
      result.push({ label: 'Record Fields', items })
    }
  }

  // Old values (only for validation_rule/when_expression)
  if (
    props.fields.length > 0 &&
    (props.context === 'validation_rule' || props.context === 'when_expression')
  ) {
    const items = props.fields
      .map((f) => ({
        label: `old.${f.apiName}`,
        value: `old.${f.apiName}`,
        detail: f.label,
      }))
      .filter((i) => !q || i.label.toLowerCase().includes(q) || (i.detail?.toLowerCase().includes(q) ?? false))
    if (items.length > 0) {
      result.push({ label: 'Old Values', items })
    }
  }

  // User variables
  if (props.context !== 'function_body') {
    const userItems: PickerItem[] = [
      { label: 'user.id', value: 'user.id', detail: 'User ID' },
      { label: 'user.profile_id', value: 'user.profile_id', detail: 'Profile ID' },
      { label: 'user.role_id', value: 'user.role_id', detail: 'Role ID' },
    ].filter((i) => !q || i.label.toLowerCase().includes(q))
    if (userItems.length > 0) {
      result.push({ label: 'User', items: userItems })
    }
  }

  // System variables
  if (props.context !== 'function_body') {
    const sysItems: PickerItem[] = [
      { label: 'now', value: 'now', detail: 'Current time' },
    ].filter((i) => !q || i.label.toLowerCase().includes(q))
    if (sysItems.length > 0) {
      result.push({ label: 'System', items: sysItems })
    }
  }

  // Parameters (for function_body context)
  if (props.context === 'function_body' && props.params.length > 0) {
    const items = props.params
      .map((p) => ({
        label: p.name,
        value: p.name,
        detail: p.type || 'any',
      }))
      .filter((i) => !q || i.label.toLowerCase().includes(q))
    if (items.length > 0) {
      result.push({ label: 'Parameters', items })
    }
  }

  return result
})

function onInsert(value: string) {
  emit('insert', value)
}
</script>

<template>
  <div class="w-60 border-l pl-3 space-y-3" data-testid="field-picker">
    <Input
      v-model="search"
      placeholder="Search..."
      class="h-8 text-xs"
      data-testid="field-picker-search"
    />

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

    <div v-if="groups.length === 0" class="text-xs text-muted-foreground">
      No available variables
    </div>
  </div>
</template>
