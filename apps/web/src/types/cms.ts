export interface ContentType {
  ID: string
  DocumentID: string
  Name: string
  Slug: string
  Kind: 'single' | 'collection'
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
