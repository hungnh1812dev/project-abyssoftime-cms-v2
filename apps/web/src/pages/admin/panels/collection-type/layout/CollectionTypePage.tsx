import { useParams } from 'react-router-dom'
import { useContentTypes } from '@/hooks/useContentTypes'
import { CollectionListPage } from './CollectionListPage'

export function CollectionTypePage() {
  const { slug } = useParams<{ slug: string }>()
  const { data: contentTypes = [], isLoading } = useContentTypes()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  const ct = contentTypes.find((c) => c.Slug === slug)

  if (!ct) {
    return <p className="text-muted-foreground">Content type "{slug}" not found.</p>
  }

  return <CollectionListPage contentType={ct} />
}
