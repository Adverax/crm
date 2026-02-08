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
  text: 'Текст',
  number: 'Число',
  boolean: 'Логический',
  datetime: 'Дата/Время',
  picklist: 'Список выбора',
  reference: 'Связь',
}

export const FIELD_SUBTYPE_LABELS: Record<FieldSubtype, string> = {
  plain: 'Простой текст',
  area: 'Многострочный текст',
  rich: 'Форматированный текст',
  email: 'Email',
  phone: 'Телефон',
  url: 'URL',
  integer: 'Целое число',
  decimal: 'Десятичное',
  currency: 'Валюта',
  percent: 'Процент',
  auto_number: 'Авто-номер',
  date: 'Дата',
  datetime: 'Дата и время',
  time: 'Время',
  single: 'Одиночный выбор',
  multi: 'Множественный выбор',
  association: 'Ассоциация',
  composition: 'Композиция',
  polymorphic: 'Полиморфная',
}

export interface ConfigFieldDef {
  key: string
  label: string
  type: 'number' | 'text' | 'boolean' | 'select'
  options?: { value: string; label: string }[]
}

export const CONFIG_FIELDS_BY_TYPE: Record<string, ConfigFieldDef[]> = {
  'text/plain': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'text/area': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'text/rich': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'text/email': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
  ],
  'text/phone': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
  ],
  'text/url': [
    { key: 'maxLength', label: 'Макс. длина', type: 'number' },
  ],
  'number/integer': [
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'number/decimal': [
    { key: 'precision', label: 'Точность (всего цифр)', type: 'number' },
    { key: 'scale', label: 'Масштаб (после запятой)', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'number/currency': [
    { key: 'precision', label: 'Точность (всего цифр)', type: 'number' },
    { key: 'scale', label: 'Масштаб (после запятой)', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'number/percent': [
    { key: 'precision', label: 'Точность (всего цифр)', type: 'number' },
    { key: 'scale', label: 'Масштаб (после запятой)', type: 'number' },
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'number/auto_number': [
    { key: 'format', label: 'Формат (например, INV-{0000})', type: 'text' },
    { key: 'startValue', label: 'Начальное значение', type: 'number' },
  ],
  'boolean': [
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'boolean' },
  ],
  'datetime/date': [
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'datetime/datetime': [
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'datetime/time': [
    { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' },
  ],
  'reference/association': [
    { key: 'relationshipName', label: 'Имя связи', type: 'text' },
    {
      key: 'onDelete', label: 'При удалении', type: 'select',
      options: [
        { value: 'set_null', label: 'Очистить (set null)' },
        { value: 'restrict', label: 'Запретить (restrict)' },
      ],
    },
  ],
  'reference/composition': [
    { key: 'relationshipName', label: 'Имя связи', type: 'text' },
    {
      key: 'onDelete', label: 'При удалении', type: 'select',
      options: [
        { value: 'cascade', label: 'Каскадное удаление' },
        { value: 'restrict', label: 'Запретить (restrict)' },
      ],
    },
    { key: 'isReparentable', label: 'Можно переназначить родителя', type: 'boolean' },
  ],
  'reference/polymorphic': [
    { key: 'relationshipName', label: 'Имя связи', type: 'text' },
    {
      key: 'onDelete', label: 'При удалении', type: 'select',
      options: [
        { value: 'set_null', label: 'Очистить (set null)' },
        { value: 'restrict', label: 'Запретить (restrict)' },
      ],
    },
  ],
}
