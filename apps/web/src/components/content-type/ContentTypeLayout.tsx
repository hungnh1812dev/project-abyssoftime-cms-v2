import type { ReactNode } from 'react'

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
  const defaultHeader = (
    <div className="flex items-center gap-3">
      <h1 className="text-xl font-semibold">{title}</h1>
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        {renderHeader ? renderHeader(defaultHeader) : defaultHeader}
        {!renderHeader && renderActions && (
          <div className="flex items-center gap-2">{renderActions()}</div>
        )}
      </div>
      {children}
    </div>
  )
}
