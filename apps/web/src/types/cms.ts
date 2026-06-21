export interface FieldDefinition {
  name: string
  type: string
  ext?: string[]
  fields?: FieldDefinition[]
}

export interface ContentTypeSummary {
  ID: string
  Name: string
  Slug: string
  Kind: 'single' | 'collection'
}

export interface ContentType extends ContentTypeSummary {
  Fields?: FieldDefinition[]
  listFields?: string[]
  CreatedAt: string
  UpdatedAt: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  start: number
  size: number
}

export type EntryStatus = 'draft' | 'modified' | 'published'

export interface Document {
  data: Record<string, unknown>
  status: EntryStatus
}

export const SYSTEM_FIELDS = ['documentId', 'locale', 'createdAt', 'updatedAt', 'createdBy', 'updatedBy'] as const

export function stripSystemFields(data: Record<string, unknown>): Record<string, unknown> {
  const content: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(data)) {
    if (!(SYSTEM_FIELDS as readonly string[]).includes(k)) {
      content[k] = v
    }
  }
  return content
}

export interface MediaAsset {
  ID: string
  url: string
  thumbnailUrl: string
  publicId: string
  fileName: string
  fileExt: string
  hash: string
  width: number
  height: number
  documentRef: string
  contentTypeId: string
  createdAt: string
}
