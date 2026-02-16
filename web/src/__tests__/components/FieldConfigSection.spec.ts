import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import FieldConfigSection from '@/components/admin/metadata/FieldConfigSection.vue'
import type { ConfigFieldDef } from '@/types/field-types'

describe('FieldConfigSection', () => {
  const tests = [
    {
      name: 'renders number config fields for text/plain',
      configFields: [
        { key: 'maxLength', label: 'Max length', type: 'number' as const },
        { key: 'defaultValue', label: 'Default value', type: 'text' as const },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('Max length')
        expect(wrapper.text()).toContain('Default value')
      },
    },
    {
      name: 'renders select field for onDelete',
      configFields: [
        {
          key: 'onDelete',
          label: 'On delete',
          type: 'select' as const,
          options: [
            { value: 'set_null', label: 'Set null' },
            { value: 'restrict', label: 'Restrict' },
          ],
        },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('On delete')
      },
    },
    {
      name: 'renders boolean field for isReparentable',
      configFields: [
        { key: 'isReparentable', label: 'Allow reparenting', type: 'boolean' as const },
      ] satisfies ConfigFieldDef[],
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.text()).toContain('Allow reparenting')
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
