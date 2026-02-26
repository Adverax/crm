<script setup lang="ts">
import { ref, computed } from 'vue'
import SectionCard from './SectionCard.vue'
import SectionPropertiesPanel from './SectionPropertiesPanel.vue'
import FieldPropertiesPanel from './FieldPropertiesPanel.vue'
import type { SectionConfig, LayoutFieldConfig, SharedLayout } from '@/types/layouts'
import type { SectionField } from './SectionCard.vue'

export interface OVSection {
  key: string
  label: string
  fields: SectionField[]
}

const props = defineProps<{
  sections: OVSection[]
  sectionConfig: Record<string, SectionConfig>
  fieldConfig: Record<string, LayoutFieldConfig>
  sharedLayouts: SharedLayout[]
}>()

const emit = defineEmits<{
  'update:sectionConfig': [config: Record<string, SectionConfig>]
  'update:fieldConfig': [config: Record<string, LayoutFieldConfig>]
}>()

const selectedSection = ref<string | null>(null)
const selectedField = ref<string | null>(null)

const selectedSectionData = computed(() => {
  if (!selectedSection.value) return null
  return props.sections.find((s) => s.key === selectedSection.value) ?? null
})

const selectedFieldData = computed(() => {
  if (!selectedField.value || !selectedSection.value) return null
  const section = props.sections.find((s) => s.key === selectedSection.value)
  if (!section) return null
  return section.fields.find((f) => f.apiName === selectedField.value) ?? null
})

function getSectionConfig(key: string): SectionConfig {
  return props.sectionConfig[key] ?? {}
}

function getFieldConfig(name: string): LayoutFieldConfig {
  return props.fieldConfig[name] ?? {}
}

function selectSection(key: string) {
  selectedSection.value = key
  selectedField.value = null
}

function selectField(sectionKey: string, fieldName: string) {
  selectedSection.value = sectionKey
  selectedField.value = fieldName
}

function onSectionConfigUpdate(config: SectionConfig) {
  if (!selectedSection.value) return
  emit('update:sectionConfig', {
    ...props.sectionConfig,
    [selectedSection.value]: config,
  })
}

function onFieldConfigUpdate(config: LayoutFieldConfig) {
  if (!selectedField.value) return
  emit('update:fieldConfig', {
    ...props.fieldConfig,
    [selectedField.value]: config,
  })
}
</script>

<template>
  <div class="flex gap-4" data-testid="form-layout-tab">
    <!-- Canvas (left) -->
    <div class="flex-1 min-w-0 space-y-3">
      <template v-if="sections.length">
        <SectionCard
          v-for="section in sections"
          :key="section.key"
          :section-key="section.key"
          :label="section.label"
          :config="getSectionConfig(section.key)"
          :fields="section.fields"
          :field-configs="fieldConfig"
          :selected="selectedSection === section.key"
          :selected-field="selectedSection === section.key ? (selectedField ?? undefined) : undefined"
          @select="selectSection(section.key)"
          @select-field="selectField(section.key, $event)"
        />
      </template>
      <div v-else class="text-center py-12 text-muted-foreground" data-testid="empty-sections">
        <p class="text-sm">This Layout's Object View has no sections configured.</p>
        <p class="text-xs mt-1">Add sections in the Object View editor first.</p>
      </div>
    </div>

    <!-- Properties panel (right) -->
    <div class="w-80 shrink-0" data-testid="properties-panel">
      <FieldPropertiesPanel
        v-if="selectedField && selectedFieldData"
        :field-name="selectedField"
        :field-type="selectedFieldData.type"
        :field-label="selectedFieldData.label"
        :config="getFieldConfig(selectedField)"
        :shared-layouts="sharedLayouts"
        @update:config="onFieldConfigUpdate"
      />
      <SectionPropertiesPanel
        v-else-if="selectedSection && selectedSectionData"
        :section-key="selectedSection"
        :section-label="selectedSectionData.label"
        :config="getSectionConfig(selectedSection)"
        @update:config="onSectionConfigUpdate"
      />
      <div v-else class="text-center py-8 text-muted-foreground text-sm">
        <p>Select a section or field to edit its properties.</p>
      </div>
    </div>
  </div>
</template>
