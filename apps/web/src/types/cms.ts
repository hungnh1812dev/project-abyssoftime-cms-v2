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
  documentId: string
  contentTypeId: string
  data: Record<string, unknown>
  status: EntryStatus
  locale: string
  createdAt: string
  updatedAt: string
  createdBy: string
  updatedBy: string
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
