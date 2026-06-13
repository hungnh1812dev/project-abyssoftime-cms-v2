export interface ContentType {
  ID: string
  DocumentID: string
  Name: string
  Slug: string
  Kind: 'single' | 'collection'
  CreatedAt: string
  UpdatedAt: string
}

export interface Document {
  ID: string
  DocumentID: string
  ContentTypeID: string
  Status: 'draft' | 'published'
  Data: Record<string, unknown>
  CreatedAt: string
  UpdatedAt: string
}
