<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ref, computed } from 'vue'
import type { PortalArg } from '@/types/portals'

const props = defineProps<{
  args: PortalArg[]
}>()

const emit = defineEmits<{
  'update:args': [value: PortalArg[]]
}>()

const selectedIndex = ref<number | null>(null)

const selectedArg = computed(() =>
  selectedIndex.value !== null ? props.args[selectedIndex.value] ?? null : null,
)

const argTypes = ['string', 'int', 'float', 'bool'] as const
const nameRegex = /^[a-z][a-z0-9_]*$/

// --- Validation ---

const nameError = computed<string | null>(() => {
  if (selectedIndex.value === null || !selectedArg.value) return null
  const name = selectedArg.value.name
  if (!name) return 'Name is required'
  if (!nameRegex.test(name)) return 'Must match ^[a-z][a-z0-9_]*$'
  const isDuplicate = props.args.some(
    (a, i) => i !== selectedIndex.value && a.name === name,
  )
  if (isDuplicate) return `Duplicate name "${name}"`
  return null
})

const defaultError = computed<string | null>(() => {
  if (!selectedArg.value) return null
  const val = selectedArg.value.default
  if (val === undefined || val === null || val === '') return null
  const type = selectedArg.value.type
  if (type === 'int' && !/^-?\d+$/.test(val)) return `"${val}" is not a valid int`
  if (type === 'float' && isNaN(Number(val))) return `"${val}" is not a valid float`
  if (type === 'bool' && val !== 'true' && val !== 'false')
    return `"${val}" is not a valid bool (use "true" or "false")`
  return null
})

const hasErrors = computed(() => {
  for (let i = 0; i < props.args.length; i++) {
    if (hasArgError(i)) return true
  }
  return false
})

function hasArgError(index: number): boolean {
  const a = props.args[index]
  if (!a) return false
  if (!a.name || !nameRegex.test(a.name)) return true
  if (props.args.some((b, j) => j !== index && b.name === a.name)) return true
  const d = a.default
  if (d != null && d !== '') {
    if (a.type === 'int' && !/^-?\d+$/.test(d)) return true
    if (a.type === 'float' && isNaN(Number(d))) return true
    if (a.type === 'bool' && d !== 'true' && d !== 'false') return true
  }
  return false
}

// expose hasErrors so parent can check before save
defineExpose({ hasErrors })

// --- Actions ---

function selectArg(index: number) {
  selectedIndex.value = index
}

function addArg() {
  const updated: PortalArg[] = [...props.args, { name: '', type: 'string' }]
  emit('update:args', updated)
  selectedIndex.value = updated.length - 1
}

function removeArg(index: number) {
  const updated = [...props.args]
  updated.splice(index, 1)
  emit('update:args', updated)
  if (selectedIndex.value === index) {
    selectedIndex.value = updated.length > 0 ? Math.min(index, updated.length - 1) : null
  } else if (selectedIndex.value !== null && selectedIndex.value > index) {
    selectedIndex.value--
  }
}

function argLabel(arg: PortalArg): string {
  return arg.name || 'unnamed'
}

function isRequired(arg: PortalArg): boolean {
  return arg.default === undefined || arg.default === null
}

function updateDefault(value: string) {
  if (!selectedArg.value) return
  if (value === '') {
    selectedArg.value.default = undefined
  } else {
    selectedArg.value.default = value
  }
}

</script>

<template>
  <div class="flex gap-4 min-h-[400px]" data-testid="args-master-detail">
    <!-- Left panel: arg list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Args</span>
        <IconButton
          :icon="Plus"
          tooltip="Add arg"
          variant="outline"
          size="sm"
          data-testid="add-arg-btn"
          @click="addArg"
        />
      </div>
      <div v-if="args.length === 0" class="p-3 text-sm text-muted-foreground">
        No args configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="(arg, aIdx) in args"
          :key="aIdx"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{
            'bg-muted': selectedIndex === aIdx,
            'border-l-2 border-l-destructive': hasArgError(aIdx),
          }"
          data-testid="arg-card"
          @click="selectArg(aIdx)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium font-mono truncate">{{ argLabel(arg) }}</div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <Badge variant="secondary" class="text-[10px]">
                {{ arg.type }}
              </Badge>
              <Badge
                :variant="isRequired(arg) ? 'destructive' : 'outline'"
                class="text-[10px]"
              >
                {{ isRequired(arg) ? 'required' : 'optional' }}
              </Badge>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: arg detail -->
    <div class="flex-1 min-w-0">
      <div
        v-if="!selectedArg"
        class="flex items-center justify-center h-full text-sm text-muted-foreground"
      >
        Select an arg to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Arg Details</span>
          <IconButton
            :icon="Trash2"
            tooltip="Delete arg"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            data-testid="delete-arg-btn"
            @click="removeArg(selectedIndex!)"
          />
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Name</Label>
          <Input
            v-model="selectedArg.name"
            placeholder="param_name"
            class="font-mono"
            :class="{ 'border-destructive': nameError }"
            data-testid="arg-name-input"
          />
          <p v-if="nameError" class="text-[11px] text-destructive" data-testid="arg-name-error">
            {{ nameError }}
          </p>
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Type</Label>
          <Select v-model="selectedArg.type" data-testid="arg-type-select">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="t in argTypes" :key="t" :value="t">
                {{ t }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="space-y-1">
          <Label class="text-xs">Default</Label>
          <p class="text-[11px] text-muted-foreground">
            Leave empty to make the arg required.
          </p>
          <Input
            :model-value="selectedArg.default ?? ''"
            placeholder="(required if empty)"
            :class="{ 'border-destructive': defaultError }"
            data-testid="arg-default-input"
            @update:model-value="updateDefault(String($event))"
          />
          <p
            v-if="defaultError"
            class="text-[11px] text-destructive"
            data-testid="arg-default-error"
          >
            {{ defaultError }}
          </p>
        </div>

      </div>
    </div>
  </div>
</template>
