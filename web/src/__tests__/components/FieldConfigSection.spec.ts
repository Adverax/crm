import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import FieldConfigSection from '@/components/admin/metadata/FieldConfigSection.vue'
import type { ConfigFieldDef } from '@/types/field-types'

describe('FieldConfigSection', () => {
  const tests = [
    {
      name: 'renders number config fields for text/plain',
      configFields: [
        { key: 'maxLength', label: 'Макс. длина', type: 'number' as const },
        { key: 'defaultValue', label: 'Значение по умолчанию', type: 'text' as const },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('Макс. длина')
        expect(wrapper.text()).toContain('Значение по умолчанию')
      },
    },
    {
      name: 'renders select field for onDelete',
      configFields: [
        {
          key: 'onDelete',
          label: 'При удалении',
          type: 'select' as const,
          options: [
            { value: 'set_null', label: 'Очистить' },
            { value: 'restrict', label: 'Запретить' },
          ],
        },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('При удалении')
      },
    },
    {
      name: 'renders boolean field for isReparentable',
      configFields: [
        { key: 'isReparentable', label: 'Можно переназначить родителя', type: 'boolean' as const },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('Можно переназначить родителя')
      },
    },
    {
      name: 'renders empty when no config fields',
      configFields: [] as ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        const inputs = wrapper.findAll('input')
        expect(inputs).toHaveLength(0)
      },
    },
  ]

  tests.forEach((tt) => {
    it(tt.name, () => {
      const wrapper = mount(FieldConfigSection, {
        props: {
          configFields: tt.configFields,
          modelValue: {},
        },
        global: {
          stubs: {
            teleport: true,
          },
        },
      })

      tt.assert(wrapper)
    })
  })
})
