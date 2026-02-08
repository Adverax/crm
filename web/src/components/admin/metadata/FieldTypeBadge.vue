<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import type { FieldType, FieldSubtype } from '@/types/metadata'
import { FIELD_TYPE_LABELS, FIELD_SUBTYPE_LABELS } from '@/types/field-types'

const props = defineProps<{
  fieldType: FieldType
  fieldSubtype?: FieldSubtype
}>()

const label = props.fieldSubtype
  ? `${FIELD_TYPE_LABELS[props.fieldType]} / ${FIELD_SUBTYPE_LABELS[props.fieldSubtype]}`
  : FIELD_TYPE_LABELS[props.fieldType]
</script>

<template>
  <Badge variant="outline">{{ label }}</Badge>
</template>
