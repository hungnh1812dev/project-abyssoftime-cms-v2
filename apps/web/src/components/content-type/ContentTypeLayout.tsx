import type { ReactNode } from 'react'
import { StickyActionBar } from '@/pages/admin/layout/StickyActionBar'

export interface ContentTypeLayoutProps {
  title: string
  status?: string
  renderHeader?: (defaultHeader: ReactNode) => ReactNode
  renderActions?: () => ReactNode
  children: ReactNode
}

export function ContentTypeLayout({
  title,
  status,
  renderHeader,
  renderActions,
  children,
}: ContentTypeLayoutProps) {
  if (renderHeader) {
    const defaultHeader = (
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold">{title}</h1>
        {status && (
          <span
            data-testid="status-badge"
            className="text-sm text-muted-foreground capitalize"
          >
            {status}
          </span>
        )}
      </div>
    )

    return (
      <div>
        <div className="sticky top-0 z-30 -mx-6 mb-6 flex h-16 items-center justify-between border-b border-border bg-background/80 px-6 backdrop-blur-sm">
          {renderHeader(defaultHeader)}
        </div>
        {children}
      </div>
    )
  }

  return (
    <div>
      <StickyActionBar title={title} status={status} renderActions={renderActions} />
      {children}
    </div>
  )
}
