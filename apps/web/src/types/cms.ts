export interface FieldDefinition {
  name: string
  type: string
  ext?: string[]
  fields?: FieldDefinition[]
}

export interface ContentType {
  ID: string
  DocumentID: string
  Name: string
  Slug: string
  Kind: 'single' | 'collection'
  Fields?: FieldDefinition[]
  CreatedAt: string
  UpdatedAt: string
}

export type EntryStatus = 'draft' | 'modified' | 'published'

export interface Document {
  EntryID: string
  ContentTypeID: string
  Data: Record<string, unknown>
  Status: EntryStatus
  Locale: string
  CreatedAt: string
  UpdatedAt: string
  CreatedBy: string
  UpdatedBy: string
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
