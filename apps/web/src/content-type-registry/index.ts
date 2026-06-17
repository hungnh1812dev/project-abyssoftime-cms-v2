import type { ComponentType } from 'react'
import type { ContentTypeLayoutProps } from '@/components/content-type/ContentTypeLayout'

export interface CollectionColumnDef {
  key: string
  label: string
  type: 'text' | 'boolean' | 'number' | 'image'
}

export interface ContentTypeRegistration {
  slug: string
  kind: 'single' | 'collection'
  columns?: CollectionColumnDef[]
  wrapper?: ComponentType<ContentTypeLayoutProps>
}

export const contentTypeRegistry: ContentTypeRegistration[] = []

export function getRegistration(slug: string): ContentTypeRegistration | undefined {
  return contentTypeRegistry.find((r) => r.slug === slug)
}
