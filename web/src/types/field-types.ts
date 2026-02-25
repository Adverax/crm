import type { FieldType, FieldSubtype } from './metadata'

export const FIELD_TYPE_SUBTYPES: Record<FieldType, FieldSubtype[]> = {
  text: ['plain', 'area', 'rich', 'email', 'phone', 'url'],
  number: ['integer', 'decimal', 'currency', 'percent', 'auto_number'],
  boolean: [],
  datetime: ['date', 'datetime', 'time'],
  picklist: ['single', 'multi'],
  reference: ['association', 'composition', 'polymorphic'],
}

export const FIELD_TYPE_LABELS: Record<FieldType, string> = {
  text: 'Text',
  number: 'Number',
  boolean: 'Boolean',
  datetime: 'Date/Time',
  picklist: 'Picklist',
  reference: 'Reference',
}

export const FIELD_SUBTYPE_LABELS: Record<FieldSubtype, string> = {
  plain: 'Plain Text',
  area: 'Text Area',
  rich: 'Rich Text',
  email: 'Email',
  phone: 'Phone',
  url: 'URL',
  integer: 'Integer',
  decimal: 'Decimal',
  currency: 'Currency',
  percent: 'Percent',
  auto_number: 'Auto Number',
  date: 'Date',
  datetime: 'Date & Time',
  time: 'Time',
  single: 'Single Select',
  multi: 'Multi Select',
  association: 'Association',
  composition: 'Composition',
  polymorphic: 'Polymorphic',
}

export interface ConfigFieldDef {
  key: string
  label: string
  type: 'number' | 'text' | 'boolean' | 'select'
  options?: { value: string; label: string }[]
}

export const CONFIG_FIELDS_BY_TYPE: Record<string, ConfigFieldDef[]> = {
  'text/plain': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'text/area': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'text/rich': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'text/email': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
  ],
  'text/phone': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
  ],
  'text/url': [
    { key: 'maxLength', label: 'Max Length', type: 'number' },
  ],
  'number/integer': [
    { key: 'precision', label: 'Precision (total digits)', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'number/decimal': [
    { key: 'precision', label: 'Precision (total digits)', type: 'number' },
    { key: 'scale', label: 'Scale (decimal places)', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'number/currency': [
    { key: 'precision', label: 'Precision (total digits)', type: 'number' },
    { key: 'scale', label: 'Scale (decimal places)', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'number/percent': [
    { key: 'precision', label: 'Precision (total digits)', type: 'number' },
    { key: 'scale', label: 'Scale (decimal places)', type: 'number' },
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'number/auto_number': [
    { key: 'format', label: 'Format (e.g., INV-{0000})', type: 'text' },
    { key: 'startValue', label: 'Start Value', type: 'number' },
  ],
  'boolean': [
    { key: 'defaultValue', label: 'Default Value', type: 'boolean' },
  ],
  'datetime/date': [
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'datetime/datetime': [
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'datetime/time': [
    { key: 'defaultValue', label: 'Default Value', type: 'text' },
  ],
  'reference/association': [
    { key: 'relationshipName', label: 'Relationship Name', type: 'text' },
    {
      key: 'onDelete', label: 'On Delete', type: 'select',
      options: [
        { value: 'set_null', label: 'Set Null' },
        { value: 'restrict', label: 'Restrict' },
      ],
    },
  ],
  'reference/composition': [
    { key: 'relationshipName', label: 'Relationship Name', type: 'text' },
    {
      key: 'onDelete', label: 'On Delete', type: 'select',
      options: [
        { value: 'cascade', label: 'Cascade Delete' },
        { value: 'restrict', label: 'Restrict' },
      ],
    },
    { key: 'isReparentable', label: 'Allow Reparenting', type: 'boolean' },
  ],
  'reference/polymorphic': [
    { key: 'relationshipName', label: 'Relationship Name', type: 'text' },
    {
      key: 'onDelete', label: 'On Delete', type: 'select',
      options: [
        { value: 'set_null', label: 'Set Null' },
        { value: 'restrict', label: 'Restrict' },
      ],
    },
  ],
}
