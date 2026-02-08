import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'

describe('ConfirmDialog', () => {
  const tests = [
    {
      name: 'accepts required props without error',
      props: {
        open: true,
        title: 'Удалить объект?',
        description: 'Это действие нельзя отменить.',
      },
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.props('title')).toBe('Удалить объект?')
        expect(wrapper.props('description')).toBe('Это действие нельзя отменить.')
      },
    },
    {
      name: 'accepts custom confirm label',
      props: {
        open: true,
        title: 'Test',
        description: 'Description',
        confirmLabel: 'Подтвердить',
      },
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.props('confirmLabel')).toBe('Подтвердить')
      },
    },
    {
      name: 'defaults confirmLabel to undefined (component uses "Удалить")',
      props: {
        open: false,
        title: 'Test',
        description: 'Description',
      },
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.props('confirmLabel')).toBeUndefined()
      },
    },
    {
      name: 'accepts loading prop',
      props: {
        open: true,
        title: 'Test',
        description: 'Description',
        loading: true,
      },
      assert: (wrapper: ReturnType<typeof mount>) => {
        expect(wrapper.props('loading')).toBe(true)
      },
    },
  ]

  tests.forEach((tt) => {
    it(tt.name, () => {
      const wrapper = mount(ConfirmDialog, {
        props: tt.props,
        shallow: true,
      })

      tt.assert(wrapper)
      wrapper.unmount()
    })
  })

  it('mounts without errors when closed', () => {
    const wrapper = mount(ConfirmDialog, {
      props: {
        open: false,
        title: 'Test',
        description: 'Desc',
      },
      shallow: true,
    })

    expect(wrapper.exists()).toBe(true)
    wrapper.unmount()
  })
})
