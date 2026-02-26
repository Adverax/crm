<script setup lang="ts">
import { computed } from 'vue'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { LayoutFieldConfig, SharedLayout } from '@/types/layouts'

const props = defineProps<{
  fieldName: string
  fieldType: string
  fieldLabel: string
  config: LayoutFieldConfig
  sharedLayouts: SharedLayout[]
}>()

const emit = defineEmits<{
  'update:config': [config: LayoutFieldConfig]
}>()

function update(patch: Partial<LayoutFieldConfig>) {
  emit('update:config', { ...props.config, ...patch })
}

const colSpan = computed({
  get: () => props.config.colSpan ?? 1,
  set: (v: number) => update({ colSpan: Math.max(1, Math.min(4, v)) }),
})

const layoutRef = computed({
  get: () => props.config.layoutRef ?? '',
  set: (v: string) => update({ layoutRef: v || undefined }),
})

const requiredExpr = computed({
  get: () => props.config.requiredExpr ?? '',
  set: (v: string) => update({ requiredExpr: v || undefined }),
})

const readonlyExpr = computed({
  get: () => props.config.readonlyExpr ?? '',
  set: (v: string) => update({ readonlyExpr: v || undefined }),
})

const visibilityExpr = computed({
  get: () => props.config.visibilityExpr ?? '',
  set: (v: string) => update({ visibilityExpr: v || undefined }),
})

const isReference = computed(() => props.fieldType === 'reference')

const refDisplayFields = computed({
  get: () => props.config.reference?.displayFields?.join(', ') ?? '',
  set: (v: string) => {
    const fields = v.split(',').map((s) => s.trim()).filter(Boolean)
    update({
      reference: {
        ...props.config.reference,
        displayFields: fields.length ? fields : undefined,
      },
    })
  },
})

const refSearchFields = computed({
  get: () => props.config.reference?.searchFields?.join(', ') ?? '',
  set: (v: string) => {
    const fields = v.split(',').map((s) => s.trim()).filter(Boolean)
    update({
      reference: {
        ...props.config.reference,
        searchFields: fields.length ? fields : undefined,
      },
    })
  },
})

const refTarget = computed({
  get: () => props.config.reference?.target ?? '',
  set: (v: string) => {
    update({
      reference: {
        ...props.config.reference,
        target: v || undefined,
      },
    })
  },
})

const refHint = computed({
  get: () => props.config.reference?.hint ?? '',
  set: (v: string) => {
    update({
      reference: {
        ...props.config.reference,
        hint: v || undefined,
      },
    })
  },
})

const fieldSharedLayouts = computed(() =>
  props.sharedLayouts.filter((sl) => sl.type === 'field'),
)
</script>

<template>
  <Card>
    <CardHeader class="py-3 px-4">
      <CardTitle class="text-sm flex items-center gap-2">
        Field: {{ fieldLabel }}
        <Badge variant="outline" class="text-xs">{{ fieldType }}</Badge>
      </CardTitle>
    </CardHeader>
    <CardContent class="px-4 pb-4 space-y-4">
      <div>
        <Label class="text-muted-foreground text-xs">API Name</Label>
        <p class="font-mono text-sm" data-testid="field-api-name">{{ fieldName }}</p>
      </div>

      <div>
        <Label for="field-col-span">Col span</Label>
        <Input
          id="field-col-span"
          type="number"
          :min="1"
          :max="4"
          :model-value="colSpan"
          data-testid="field-col-span"
          @update:model-value="colSpan = Number($event)"
        />
      </div>

      <div>
        <Label for="field-layout-ref">Shared layout ref</Label>
        <Select :model-value="layoutRef" @update:model-value="layoutRef = String($event ?? '')">
          <SelectTrigger id="field-layout-ref" data-testid="field-layout-ref">
            <SelectValue placeholder="None" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">None</SelectItem>
            <SelectItem
              v-for="sl in fieldSharedLayouts"
              :key="sl.id"
              :value="sl.apiName"
            >
              {{ sl.label }} ({{ sl.apiName }})
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label for="field-required-expr">Required expression</Label>
        <Textarea
          id="field-required-expr"
          :model-value="requiredExpr"
          placeholder="e.g. true"
          rows="2"
          class="font-mono text-sm"
          data-testid="field-required-expr"
          @update:model-value="requiredExpr = String($event)"
        />
      </div>

      <div>
        <Label for="field-readonly-expr">Readonly expression</Label>
        <Textarea
          id="field-readonly-expr"
          :model-value="readonlyExpr"
          placeholder="e.g. record.Status == 'Closed'"
          rows="2"
          class="font-mono text-sm"
          data-testid="field-readonly-expr"
          @update:model-value="readonlyExpr = String($event)"
        />
      </div>

      <div>
        <Label for="field-visibility-expr">Visibility expression</Label>
        <Textarea
          id="field-visibility-expr"
          :model-value="visibilityExpr"
          placeholder="e.g. record.Type == 'Business'"
          rows="2"
          class="font-mono text-sm"
          data-testid="field-visibility-expr"
          @update:model-value="visibilityExpr = String($event)"
        />
      </div>

      <!-- Reference config -->
      <template v-if="isReference">
        <div class="border-t pt-3">
          <p class="text-sm font-medium mb-3">Reference config</p>

          <div class="space-y-3">
            <div>
              <Label for="ref-display-fields">Display fields</Label>
              <Input
                id="ref-display-fields"
                :model-value="refDisplayFields"
                placeholder="Name, Email (comma-separated)"
                data-testid="ref-display-fields"
                @update:model-value="refDisplayFields = String($event)"
              />
            </div>

            <div>
              <Label for="ref-search-fields">Search fields</Label>
              <Input
                id="ref-search-fields"
                :model-value="refSearchFields"
                placeholder="Name (comma-separated)"
                data-testid="ref-search-fields"
                @update:model-value="refSearchFields = String($event)"
              />
            </div>

            <div>
              <Label for="ref-target">Target</Label>
              <Select :model-value="refTarget" @update:model-value="refTarget = String($event ?? '')">
                <SelectTrigger id="ref-target" data-testid="ref-target">
                  <SelectValue placeholder="Default" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">Default</SelectItem>
                  <SelectItem value="popup">Popup</SelectItem>
                  <SelectItem value="inline">Inline</SelectItem>
                  <SelectItem value="link">Link</SelectItem>
                  <SelectItem value="drawer">Drawer</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label for="ref-hint">Hint</Label>
              <Input
                id="ref-hint"
                :model-value="refHint"
                placeholder="Lookup hint text"
                data-testid="ref-hint"
                @update:model-value="refHint = String($event)"
              />
            </div>
          </div>
        </div>
      </template>
    </CardContent>
  </Card>
</template>
