import { describe, it, expect } from 'vitest'
import { useObjectForm } from '@/composables/useObjectForm'
import type { ObjectDefinition } from '@/types/metadata'

describe('useObjectForm', () => {
  describe('validate', () => {
    const tests = [
      {
        name: 'returns false when apiName is empty',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = ''
          state.label = 'Test'
          state.pluralLabel = 'Tests'
        },
        valid: false,
        errorField: 'apiName',
      },
      {
        name: 'returns false when apiName is too short',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = 'A'
          state.label = 'Test'
          state.pluralLabel = 'Tests'
        },
        valid: false,
        errorField: 'apiName',
      },
      {
        name: 'returns false when apiName has invalid characters',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = '123abc'
          state.label = 'Test'
          state.pluralLabel = 'Tests'
        },
        valid: false,
        errorField: 'apiName',
      },
      {
        name: 'returns false when label is empty',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = 'Invoice__c'
          state.label = ''
          state.pluralLabel = 'Tests'
        },
        valid: false,
        errorField: 'label',
      },
      {
        name: 'returns false when pluralLabel is empty',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = 'Invoice__c'
          state.label = 'Invoice'
          state.pluralLabel = ''
        },
        valid: false,
        errorField: 'pluralLabel',
      },
      {
        name: 'returns true when all required fields are valid',
        setup: (state: ReturnType<typeof useObjectForm>['state']) => {
          state.apiName = 'Invoice__c'
          state.label = 'Invoice'
          state.pluralLabel = 'Invoices'
        },
        valid: true,
        errorField: null,
      },
    ]

    tests.forEach((tt) => {
      it(tt.name, () => {
        const { state, errors, validate } = useObjectForm()
        tt.setup(state)

        const result = validate()

        expect(result).toBe(tt.valid)
        if (tt.errorField) {
          expect(errors[tt.errorField as keyof typeof errors]).toBeTruthy()
        }
      })
    })
  })

  describe('initFrom', () => {
    it('initializes state from existing ObjectDefinition', () => {
      const existing: ObjectDefinition = {
        id: 'uuid-1',
        apiName: 'Account',
        label: 'Аккаунт',
        pluralLabel: 'Аккаунты',
        description: 'Описание',
        objectType: 'standard',
        isPlatformManaged: false,
        isVisibleInSetup: true,
        isCustomFieldsAllowed: false,
        isDeleteableObject: false,
        isCreateable: true,
        isUpdateable: true,
        isDeleteable: false,
        isQueryable: true,
        isSearchable: true,
        hasActivities: true,
        hasNotes: true,
        hasHistoryTracking: false,
        hasSharingRules: false,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      }

      const { state, initFrom } = useObjectForm()
      initFrom(existing)

      expect(state.apiName).toBe('Account')
      expect(state.label).toBe('Аккаунт')
      expect(state.objectType).toBe('standard')
      expect(state.isSearchable).toBe(true)
      expect(state.hasActivities).toBe(true)
    })
  })

  describe('toCreateRequest', () => {
    it('returns a valid CreateObjectRequest', () => {
      const { state, toCreateRequest } = useObjectForm()
      state.apiName = 'Invoice__c'
      state.label = 'Invoice'
      state.pluralLabel = 'Invoices'
      state.objectType = 'custom'

      const req = toCreateRequest()

      expect(req.apiName).toBe('Invoice__c')
      expect(req.label).toBe('Invoice')
      expect(req.pluralLabel).toBe('Invoices')
      expect(req.objectType).toBe('custom')
    })
  })

  describe('toUpdateRequest', () => {
    it('returns a valid UpdateObjectRequest without apiName or objectType', () => {
      const { state, toUpdateRequest } = useObjectForm()
      state.apiName = 'Invoice__c'
      state.label = 'Updated'
      state.pluralLabel = 'Updated items'
      state.objectType = 'custom'

      const req = toUpdateRequest()

      expect(req.label).toBe('Updated')
      expect(req.pluralLabel).toBe('Updated items')
      expect((req as Record<string, unknown>)['apiName']).toBeUndefined()
      expect((req as Record<string, unknown>)['objectType']).toBeUndefined()
    })
  })

  describe('reset', () => {
    it('resets state to defaults', () => {
      const { state, reset } = useObjectForm()
      state.apiName = 'Something'
      state.label = 'Something'
      state.isSearchable = true

      reset()

      expect(state.apiName).toBe('')
      expect(state.label).toBe('')
      expect(state.isSearchable).toBe(false)
    })
  })
})
