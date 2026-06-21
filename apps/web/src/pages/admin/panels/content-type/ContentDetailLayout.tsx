import type { ReactNode } from 'react'
import { StickyActionBar } from '@/pages/admin/layout/StickyActionBar'

interface Props {
  title: string
  status?: string
  backLink?: ReactNode
  renderActions?: () => ReactNode
  children: ReactNode
}

export function ContentDetailLayout({ title, status, backLink, renderActions, children }: Props) {
  return (
    <div>
      <StickyActionBar title={title} status={status} renderActions={renderActions} />
      {backLink && <div className="mb-4">{backLink}</div>}
      {children}
    </div>
  )
}
