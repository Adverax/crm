<script setup lang="ts">
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Plus } from 'lucide-vue-next'

defineEmits<{
  select: [type: string]
}>()

const groups = [
  {
    label: 'Record',
    items: [
      { type: 'record.create', label: 'Create Record' },
      { type: 'record.update', label: 'Update Record' },
      { type: 'record.delete', label: 'Delete Record' },
      { type: 'record.get', label: 'Get Record' },
      { type: 'record.query', label: 'Query Records' },
    ],
  },
  {
    label: 'Compute',
    items: [
      { type: 'compute.transform', label: 'Transform Data' },
      { type: 'compute.validate', label: 'Validate' },
      { type: 'compute.fail', label: 'Fail' },
    ],
  },
  {
    label: 'Flow',
    items: [
      { type: 'flow.if', label: 'If / Else' },
      { type: 'flow.match', label: 'Match / Switch' },
      { type: 'flow.call', label: 'Call Procedure' },
      { type: 'flow.try', label: 'Try / Catch' },
    ],
  },
  {
    label: 'Integration',
    items: [
      { type: 'integration.http', label: 'HTTP Request' },
    ],
  },
  {
    label: 'Notification',
    items: [
      { type: 'notification.email', label: 'Send Email (stub)' },
    ],
  },
  {
    label: 'Wait',
    items: [
      { type: 'wait.delay', label: 'Delay (stub)' },
    ],
  },
]
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="outline" class="w-full" data-testid="add-command-btn">
        <Plus class="h-4 w-4 mr-2" />
        Add Command
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-56">
      <template v-for="(group, gi) in groups" :key="group.label">
        <DropdownMenuSeparator v-if="gi > 0" />
        <DropdownMenuLabel>{{ group.label }}</DropdownMenuLabel>
        <DropdownMenuGroup>
          <DropdownMenuItem
            v-for="item in group.items"
            :key="item.type"
            :data-testid="`cmd-${item.type}`"
            @click="$emit('select', item.type)"
          >
            {{ item.label }}
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </template>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
