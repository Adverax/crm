<script setup lang="ts">
import { computed } from 'vue'
import CommandEditor from './CommandEditor.vue'
import CommandPicker from './CommandPicker.vue'
import { Card, CardContent } from '@/components/ui/card'
import type { CommandDef } from '@/types/procedures'

const props = defineProps<{
  commands: CommandDef[]
}>()

const emit = defineEmits<{
  'update:commands': [commands: CommandDef[]]
}>()

function addCommand(type: string) {
  const newCmd: CommandDef = { type }
  emit('update:commands', [...props.commands, newCmd])
}

function updateCommand(index: number, cmd: CommandDef) {
  const updated = [...props.commands]
  updated[index] = cmd
  emit('update:commands', updated)
}

function removeCommand(index: number) {
  const updated = props.commands.filter((_, i) => i !== index)
  emit('update:commands', updated)
}

function moveCommand(index: number, direction: 'up' | 'down') {
  const newIndex = direction === 'up' ? index - 1 : index + 1
  if (newIndex < 0 || newIndex >= props.commands.length) return
  const updated = [...props.commands]
  const a = updated[index]!
  const b = updated[newIndex]!
  updated[index] = b
  updated[newIndex] = a
  emit('update:commands', updated)
}

const isEmpty = computed(() => props.commands.length === 0)
</script>

<template>
  <div class="space-y-4">
    <div v-if="isEmpty" class="text-center py-8 text-muted-foreground">
      <p>No commands yet. Add your first command to start building the procedure.</p>
    </div>

    <div v-else class="space-y-2">
      <div v-for="(cmd, index) in commands" :key="index">
        <CommandEditor
          :command="cmd"
          :index="index"
          :total="commands.length"
          @update="updateCommand(index, $event)"
          @remove="removeCommand(index)"
          @move-up="moveCommand(index, 'up')"
          @move-down="moveCommand(index, 'down')"
        />

        <!-- flow.if: Then / Else blocks -->
        <template v-if="cmd.type === 'flow.if'">
          <div class="ml-6 mt-2 space-y-2">
            <div class="text-xs font-medium text-purple-600">Then</div>
            <ProcedureConstructor
              :commands="cmd.then ?? []"
              @update:commands="updateCommand(index, { ...cmd, then: $event })"
            />
            <div class="text-xs font-medium text-purple-600">Else</div>
            <ProcedureConstructor
              :commands="cmd.else ?? []"
              @update:commands="updateCommand(index, { ...cmd, else: $event })"
            />
          </div>
        </template>

        <!-- flow.try: Try / Catch blocks -->
        <template v-if="cmd.type === 'flow.try'">
          <div class="ml-6 mt-2 space-y-2">
            <div class="text-xs font-medium text-purple-600" data-testid="try-block-label">Try</div>
            <ProcedureConstructor
              :commands="cmd.try ?? []"
              @update:commands="updateCommand(index, { ...cmd, try: $event })"
            />
            <div class="text-xs font-medium text-purple-600" data-testid="catch-block-label">Catch</div>
            <ProcedureConstructor
              :commands="cmd.catch ?? []"
              @update:commands="updateCommand(index, { ...cmd, catch: $event })"
            />
          </div>
        </template>
      </div>
    </div>

    <Card>
      <CardContent class="py-3">
        <CommandPicker @select="addCommand" />
      </CardContent>
    </Card>
  </div>
</template>
