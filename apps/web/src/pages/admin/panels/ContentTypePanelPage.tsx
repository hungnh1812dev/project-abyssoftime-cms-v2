import { useParams } from 'react-router-dom'
import { useContentTypes } from '@/hooks/useContentTypes'
import { SingleTypePanel } from './SingleTypePanel'

export function ContentTypePanelPage() {
  const { slug } = useParams<{ slug: string }>()
  const { data: contentTypes = [], isLoading } = useContentTypes()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  const ct = contentTypes.find((c) => c.Slug === slug)

  if (!ct) {
    return <p className="text-muted-foreground">Content type "{slug}" not found.</p>
  }

  if (ct.Kind === 'single') {
    return <SingleTypePanel contentType={ct} />
  }

  return <p className="text-muted-foreground">Collection panels coming in T3.10.</p>
}
