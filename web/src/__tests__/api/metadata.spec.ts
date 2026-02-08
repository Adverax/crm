import { describe, it, expect, vi, beforeEach } from 'vitest'
import { metadataApi } from '@/api/metadata'
import { http } from '@/api/http'
import type { CreateObjectRequest, UpdateObjectRequest, CreateFieldRequest, UpdateFieldRequest } from '@/types/metadata'

vi.mock('@/api/http', () => ({
  http: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

const mockedHttp = vi.mocked(http)

describe('metadataApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listObjects', () => {
    const tests = [
      {
        name: 'calls GET /api/v1/admin/metadata/objects without params',
        filter: undefined,
        expectedParams: {},
      },
      {
        name: 'passes page and perPage params',
        filter: { page: 2, perPage: 10 },
        expectedParams: { page: 2, per_page: 10 },
      },
      {
        name: 'passes objectType filter',
        filter: { objectType: 'custom' as const },
        expectedParams: { object_type: 'custom' },
      },
    ]

    tests.forEach((tt) => {
      it(tt.name, async () => {
        const mockResponse = { data: [], pagination: { page: 1, perPage: 20, total: 0, totalPages: 0 } }
        mockedHttp.get.mockResolvedValue(mockResponse)

        await metadataApi.listObjects(tt.filter)

        expect(mockedHttp.get).toHaveBeenCalledWith(
          '/api/v1/admin/metadata/objects',
          expect.objectContaining(tt.expectedParams),
        )
      })
    })
  })

  describe('getObject', () => {
    it('calls GET with object ID', async () => {
      const mockResponse = { data: { id: 'abc-123', apiName: 'Account' } }
      mockedHttp.get.mockResolvedValue(mockResponse)

      await metadataApi.getObject('abc-123')

      expect(mockedHttp.get).toHaveBeenCalledWith('/api/v1/admin/metadata/objects/abc-123')
    })
  })

  describe('createObject', () => {
    it('calls POST with request body', async () => {
      const req: CreateObjectRequest = {
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
      }
      const mockResponse = { data: { id: 'new-id', ...req } }
      mockedHttp.post.mockResolvedValue(mockResponse)

      await metadataApi.createObject(req)

      expect(mockedHttp.post).toHaveBeenCalledWith('/api/v1/admin/metadata/objects', req)
    })
  })

  describe('updateObject', () => {
    it('calls PUT with object ID and body', async () => {
      const req: UpdateObjectRequest = { label: 'Updated', pluralLabel: 'Updated items' }
      mockedHttp.put.mockResolvedValue({ data: {} })

      await metadataApi.updateObject('obj-1', req)

      expect(mockedHttp.put).toHaveBeenCalledWith('/api/v1/admin/metadata/objects/obj-1', req)
    })
  })

  describe('deleteObject', () => {
    it('calls DELETE with object ID', async () => {
      mockedHttp.delete.mockResolvedValue(undefined)

      await metadataApi.deleteObject('obj-1')

      expect(mockedHttp.delete).toHaveBeenCalledWith('/api/v1/admin/metadata/objects/obj-1')
    })
  })

  describe('listFields', () => {
    it('calls GET with object ID', async () => {
      mockedHttp.get.mockResolvedValue({ data: [] })

      await metadataApi.listFields('obj-1')

      expect(mockedHttp.get).toHaveBeenCalledWith('/api/v1/admin/metadata/objects/obj-1/fields')
    })
  })

  describe('createField', () => {
    it('calls POST with object ID and field data', async () => {
      const req: CreateFieldRequest = {
        apiName: 'amount__c',
        label: 'Amount',
        fieldType: 'number',
        fieldSubtype: 'currency',
      }
      mockedHttp.post.mockResolvedValue({ data: { id: 'field-1', ...req } })

      await metadataApi.createField('obj-1', req)

      expect(mockedHttp.post).toHaveBeenCalledWith('/api/v1/admin/metadata/objects/obj-1/fields', req)
    })
  })

  describe('updateField', () => {
    it('calls PUT with object ID, field ID, and data', async () => {
      const req: UpdateFieldRequest = { label: 'Updated Amount' }
      mockedHttp.put.mockResolvedValue({ data: {} })

      await metadataApi.updateField('obj-1', 'field-1', req)

      expect(mockedHttp.put).toHaveBeenCalledWith(
        '/api/v1/admin/metadata/objects/obj-1/fields/field-1',
        req,
      )
    })
  })

  describe('deleteField', () => {
    it('calls DELETE with object ID and field ID', async () => {
      mockedHttp.delete.mockResolvedValue(undefined)

      await metadataApi.deleteField('obj-1', 'field-1')

      expect(mockedHttp.delete).toHaveBeenCalledWith(
        '/api/v1/admin/metadata/objects/obj-1/fields/field-1',
      )
    })
  })
})
