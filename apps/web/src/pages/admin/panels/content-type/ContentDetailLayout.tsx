import type { ReactNode } from 'react'
import { StickyActionBar } from '@/pages/admin/layout/StickyActionBar'

interface Props {
  title: string
  status?: string
  backLink?: ReactNode
  metadata?: ReactNode
  renderActions?: () => ReactNode
  children: ReactNode
}

export function ContentDetailLayout({ title, status, backLink, metadata, renderActions, children }: Props) {
  return (
    <div className="min-h-full">
      <StickyActionBar title={title} status={status} renderActions={renderActions} />
      <div className="p-6">
        {backLink && <div className="mb-4">{backLink}</div>}
        {metadata && <div className="mb-4">{metadata}</div>}
        {children}
      </div>
    </div>
  )
}
