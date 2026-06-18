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
  CreatedAt: string
  UpdatedAt: string
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
  documentRef: string
  contentTypeId: string
  createdAt: string
}
