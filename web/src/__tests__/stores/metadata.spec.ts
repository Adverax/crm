import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMetadataStore } from '@/stores/metadata'
import { metadataApi } from '@/api/metadata'
import type { ObjectDefinition, FieldDefinition } from '@/types/metadata'

vi.mock('@/api/metadata', () => ({
  metadataApi: {
    listObjects: vi.fn(),
    getObject: vi.fn(),
    createObject: vi.fn(),
    updateObject: vi.fn(),
    deleteObject: vi.fn(),
    listFields: vi.fn(),
    createField: vi.fn(),
    updateField: vi.fn(),
    deleteField: vi.fn(),
  },
}))

const mockedApi = vi.mocked(metadataApi)

const fakeObject: ObjectDefinition = {
  id: 'obj-1',
  apiName: 'Invoice__c',
  label: 'Invoice',
  pluralLabel: 'Invoices',
  description: '',
  objectType: 'custom',
  isPlatformManaged: false,
  isVisibleInSetup: true,
  isCustomFieldsAllowed: true,
  isDeleteableObject: true,
  isCreateable: true,
  isUpdateable: true,
  isDeleteable: true,
  isQueryable: true,
  isSearchable: false,
  hasActivities: false,
  hasNotes: false,
  hasHistoryTracking: false,
  hasSharingRules: false,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const fakeField: FieldDefinition = {
  id: 'field-1',
  objectId: 'obj-1',
  apiName: 'amount__c',
  label: 'Amount',
  description: '',
  helpText: '',
  fieldType: 'number',
  fieldSubtype: 'currency',
  isRequired: true,
  isUnique: false,
  config: { precision: 18, scale: 2 },
  isSystemField: false,
  isCustom: true,
  isPlatformManaged: false,
  sortOrder: 0,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

describe('useMetadataStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('fetchObjects', () => {
    const tests = [
      {
        name: 'sets objects and pagination on success',
        setup: () => {
          mockedApi.listObjects.mockResolvedValue({
            data: [fakeObject],
            pagination: { page: 1, perPage: 20, total: 1, totalPages: 1 },
          })
        },
        assert: (store: ReturnType<typeof useMetadataStore>) => {
          expect(store.objects).toHaveLength(1)
          expect(store.objects[0].apiName).toBe('Invoice__c')
          expect(store.pagination?.total).toBe(1)
          expect(store.objectsLoading).toBe(false)
          expect(store.objectsError).toBeNull()
        },
      },
      {
        name: 'sets error on failure',
        setup: () => {
          mockedApi.listObjects.mockRejectedValue(new Error('Network error'))
        },
        assert: (store: ReturnType<typeof useMetadataStore>) => {
          expect(store.objectsError).toBe('Network error')
          expect(store.objectsLoading).toBe(false)
        },
      },
    ]

    tests.forEach((tt) => {
      it(tt.name, async () => {
        tt.setup()
        const store = useMetadataStore()

        try {
          await store.fetchObjects()
        } catch {
          // expected in error case
        }

        tt.assert(store)
      })
    })
  })

  describe('fetchObject', () => {
    it('sets currentObject on success', async () => {
      mockedApi.getObject.mockResolvedValue({ data: fakeObject })
      const store = useMetadataStore()

      await store.fetchObject('obj-1')

      expect(store.currentObject?.apiName).toBe('Invoice__c')
    })
  })

  describe('createObject', () => {
    it('returns created object', async () => {
      mockedApi.createObject.mockResolvedValue({ data: fakeObject })
      const store = useMetadataStore()

      const result = await store.createObject({
        apiName: 'Invoice__c',
        label: 'Invoice',
        pluralLabel: 'Invoices',
        description: '',
        objectType: 'custom',
        isVisibleInSetup: true,
        isCustomFieldsAllowed: true,
        isDeleteableObject: true,
        isCreateable: true,
        isUpdateable: true,
        isDeleteable: true,
        isQueryable: true,
        isSearchable: false,
        hasActivities: false,
        hasNotes: false,
        hasHistoryTracking: false,
        hasSharingRules: false,
      })

      expect(result.id).toBe('obj-1')
      expect(store.objectsLoading).toBe(false)
    })
  })

  describe('deleteObject', () => {
    it('calls API and clears loading', async () => {
      mockedApi.deleteObject.mockResolvedValue(undefined)
      const store = useMetadataStore()

      await store.deleteObject('obj-1')

      expect(mockedApi.deleteObject).toHaveBeenCalledWith('obj-1')
      expect(store.objectsLoading).toBe(false)
    })
  })

  describe('fetchFields', () => {
    it('sets fields on success', async () => {
      mockedApi.listFields.mockResolvedValue({ data: [fakeField] })
      const store = useMetadataStore()

      await store.fetchFields('obj-1')

      expect(store.fields).toHaveLength(1)
      expect(store.fields[0].apiName).toBe('amount__c')
    })
  })

  describe('createField', () => {
    it('returns created field', async () => {
      mockedApi.createField.mockResolvedValue({ data: fakeField })
      const store = useMetadataStore()

      const result = await store.createField('obj-1', {
        apiName: 'amount__c',
        label: 'Amount',
        fieldType: 'number',
        fieldSubtype: 'currency',
      })

      expect(result.id).toBe('field-1')
    })
  })

  describe('updateField', () => {
    it('returns updated field', async () => {
      const updated = { ...fakeField, label: 'Updated Amount' }
      mockedApi.updateField.mockResolvedValue({ data: updated })
      const store = useMetadataStore()

      const result = await store.updateField('obj-1', 'field-1', { label: 'Updated Amount' })

      expect(result.label).toBe('Updated Amount')
    })
  })

  describe('deleteField', () => {
    it('calls API and clears loading', async () => {
      mockedApi.deleteField.mockResolvedValue(undefined)
      const store = useMetadataStore()

      await store.deleteField('obj-1', 'field-1')

      expect(mockedApi.deleteField).toHaveBeenCalledWith('obj-1', 'field-1')
      expect(store.fieldsLoading).toBe(false)
    })
  })

  describe('loading state', () => {
    it('sets objectsLoading to true during fetch and false after', async () => {
      let resolveFn: (value: unknown) => void
      mockedApi.listObjects.mockReturnValue(
        new Promise((resolve) => { resolveFn = resolve }),
      )
      const store = useMetadataStore()

      const promise = store.fetchObjects()
      expect(store.objectsLoading).toBe(true)

      resolveFn!({
        data: [],
        pagination: { page: 1, perPage: 20, total: 0, totalPages: 0 },
      })
      await promise

      expect(store.objectsLoading).toBe(false)
    })
  })
})
