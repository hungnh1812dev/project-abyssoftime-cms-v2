export interface FieldDefinition {
  name: string;
  type: string;
  ext?: string[];
  width?: '100%' | '50%' | '1/3';
  repeatable?: boolean;
  fields?: FieldDefinition[];
}

export interface ContentTypeSummary {
  ID: string;
  Name: string;
  Slug: string;
  Kind: 'single' | 'collection';
}

export interface ContentType extends ContentTypeSummary {
  Fields?: FieldDefinition[];
  listFields?: string[];
  CreatedAt: string;
  UpdatedAt: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  start: number;
  size: number;
  listFields?: string[];
}

export type EntryStatus = 'draft' | 'modified' | 'published';

export interface Document {
  data: Record<string, unknown>;
  status: EntryStatus;
}

export const SYSTEM_FIELDS = ['id', 'documentId', 'locale', 'createdAt', 'updatedAt', 'createdBy', 'updatedBy', 'updatedByName'] as const;

export function stripSystemFields(data: Record<string, unknown>): Record<string, unknown> {
  const content: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(data)) {
    if (!(SYSTEM_FIELDS as readonly string[]).includes(key)) {
      content[key] = value;
    }
  }
  return content;
}

export interface Locale {
  code: string;
  name: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface MediaAsset {
  ID: string;
  documentId: string;
  url: string;
  thumbnailUrl: string;
  publicId: string;
  fileName: string;
  fileExt: string;
  hash: string;
  width: number;
  height: number;
  createdAt: string;
}
