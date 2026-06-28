<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { ref, computed } from 'vue'
import type { PortalGate, PortalViewField, GateBodyStep, GateOutcome } from '@/types/portals'
import GateBodyStepsEditor from './GateBodyStepsEditor.vue'
import GateLayoutEditor from './GateLayoutEditor.vue'
import GateOutcomesEditor from './GateOutcomesEditor.vue'


const props = defineProps<{
  gates: Record<string, PortalGate>
  entryGate: string
}>()

const emit = defineEmits<{
  'update:gates': [value: Record<string, PortalGate>]
  'update:entryGate': [value: string]
}>()

const selectedGateName = ref<string | null>(null)

const gateNames = computed(() => Object.keys(props.gates))

const selectedGate = computed(() =>
  selectedGateName.value ? props.gates[selectedGateName.value] ?? null : null,
)

function selectGate(name: string) {
  selectedGateName.value = name
}

function addGate() {
  const baseName = 'new_gate'
  let name = baseName
  let i = 1
  while (props.gates[name]) {
    name = `${baseName}_${i++}`
  }

  const updated = {
    ...props.gates,
    [name]: {
      label: 'New Gate',
      body: [],
      outcomes: [],
    } as PortalGate,
  }
  emit('update:gates', updated)
  selectedGateName.value = name
}

function removeGate(name: string) {
  const updated = { ...props.gates }
  delete updated[name]
  emit('update:gates', updated)
  if (selectedGateName.value === name) {
    const remaining = Object.keys(updated)
    selectedGateName.value = remaining.length > 0 ? remaining[0]! : null
  }
  const keys = Object.keys(updated)
  if (props.entryGate === name && keys.length > 0) {
    emit('update:entryGate', keys[0]!)
  }
}

function renameGate(oldName: string, newName: string) {
  if (!newName || newName === oldName || props.gates[newName]) return
  const updated: Record<string, PortalGate> = {}
  for (const [key, value] of Object.entries(props.gates)) {
    updated[key === oldName ? newName : key] = value
  }
  emit('update:gates', updated)
  if (selectedGateName.value === oldName) {
    selectedGateName.value = newName
  }
  if (props.entryGate === oldName) {
    emit('update:entryGate', newName)
  }
}

function setEntryGate(name: string) {
  emit('update:entryGate', name)
}

function updateBody(steps: GateBodyStep[]) {
  if (!selectedGateName.value || !selectedGate.value) return
  const updated = { ...props.gates }
  updated[selectedGateName.value] = { ...selectedGate.value, body: steps }
  emit('update:gates', updated)
}

function updateFields(fields: PortalViewField[]) {
  if (!selectedGateName.value || !selectedGate.value) return
  const updated = { ...props.gates }
  const layout = selectedGate.value.layout ?? { fields: [] }
  updated[selectedGateName.value] = { ...selectedGate.value, layout: { ...layout, fields } }
  emit('update:gates', updated)
}

function updateOutcomes(outcomes: GateOutcome[]) {
  if (!selectedGateName.value || !selectedGate.value) return
  const updated = { ...props.gates }
  updated[selectedGateName.value] = { ...selectedGate.value, outcomes }
  emit('update:gates', updated)
}

function updateLabel(value: string) {
  if (!selectedGateName.value || !selectedGate.value) return
  const updated = { ...props.gates }
  updated[selectedGateName.value] = { ...selectedGate.value, label: value }
  emit('update:gates', updated)
}

// Rename helper ref
const editingName = ref<string | null>(null)
const editNameValue = ref('')

function startRename(name: string) {
  editingName.value = name
  editNameValue.value = name
}

function commitRename() {
  if (editingName.value && editNameValue.value) {
    renameGate(editingName.value, editNameValue.value)
  }
  editingName.value = null
}
</script>

<template>
  <div class="flex gap-4 min-h-[500px]" data-testid="gates-master-detail">
    <!-- Left panel: gate list -->
    <div class="w-64 shrink-0 border rounded-md">
      <div class="flex items-center justify-between p-3 border-b">
        <span class="text-sm font-medium">Gates</span>
        <IconButton
          :icon="Plus"
          tooltip="Add gate"
          variant="outline"
          size="sm"
          data-testid="add-gate-btn"
          @click="addGate"
        />
      </div>
      <div v-if="gateNames.length === 0" class="p-3 text-sm text-muted-foreground">
        No gates configured.
      </div>
      <div v-else class="divide-y">
        <button
          v-for="name in gateNames"
          :key="name"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-muted/50 transition-colors"
          :class="{ 'bg-muted': selectedGateName === name }"
          data-testid="gate-card"
          @click="selectGate(name)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <div class="text-sm font-medium font-mono truncate">{{ name }}</div>
              <div class="text-[11px] text-muted-foreground truncate">{{ gates[name]?.label }}</div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <Badge
                v-if="name === entryGate"
                variant="default"
                class="text-[10px]"
              >
                entry
              </Badge>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- Right panel: gate detail -->
    <div class="flex-1 min-w-0">
      <div
        v-if="!selectedGate || !selectedGateName"
        class="flex items-center justify-center h-full text-sm text-muted-foreground"
      >
        Select a gate to edit
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Gate: {{ selectedGateName }}</span>
          <div class="flex items-center gap-2">
            <button
              v-if="selectedGateName !== entryGate"
              type="button"
              class="text-xs text-primary hover:underline"
              data-testid="set-entry-btn"
              @click="setEntryGate(selectedGateName)"
            >
              Set as entry
            </button>
            <IconButton
              :icon="Trash2"
              tooltip="Delete gate"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              data-testid="delete-gate-btn"
              @click="removeGate(selectedGateName)"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1">
            <Label class="text-xs">Key</Label>
            <template v-if="editingName === selectedGateName">
              <Input
                v-model="editNameValue"
                class="font-mono"
                data-testid="gate-key-input"
                @blur="commitRename"
                @keydown.enter="commitRename"
              />
            </template>
            <template v-else>
              <div
                class="text-sm font-mono cursor-pointer hover:underline px-3 py-2 border rounded-md"
                data-testid="gate-key-display"
                @click="startRename(selectedGateName)"
              >
                {{ selectedGateName }}
              </div>
            </template>
          </div>

          <div class="space-y-1">
            <Label class="text-xs">Label</Label>
            <Input
              :model-value="selectedGate.label"
              placeholder="Gate Label"
              data-testid="gate-label-input"
              @update:model-value="updateLabel(String($event))"
            />
          </div>
        </div>

        <Tabs default-value="body" class="mt-2">
          <TabsList data-testid="gate-sub-tabs">
            <TabsTrigger value="body">Instructions</TabsTrigger>
            <TabsTrigger value="layout">Layout</TabsTrigger>
            <TabsTrigger value="outcomes">Outcomes</TabsTrigger>
          </TabsList>

          <TabsContent value="body">
            <GateBodyStepsEditor
              :steps="selectedGate.body ?? []"
              @update:steps="updateBody"
            />
          </TabsContent>

          <TabsContent value="layout">
            <GateLayoutEditor
              :fields="selectedGate.layout?.fields ?? []"
              :body-steps="selectedGate.body ?? []"
              @update:fields="updateFields"
            />
          </TabsContent>

          <TabsContent value="outcomes">
            <GateOutcomesEditor
              :outcomes="selectedGate.outcomes ?? []"
              :gate-names="gateNames"
              @update:outcomes="updateOutcomes"
            />
          </TabsContent>

        </Tabs>
      </div>
    </div>
  </div>
</template>
