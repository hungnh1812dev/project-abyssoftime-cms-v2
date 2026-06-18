import type { ReactNode } from 'react'

interface Props {
  title: string
  status?: string
  backLink?: ReactNode
  renderActions?: () => ReactNode
  children: ReactNode
}

export function ContentDetailLayout({ title, status, backLink, renderActions, children }: Props) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold">{title}</h1>
          {status && (
            <span data-testid="status-badge" className="text-sm text-muted-foreground capitalize">
              {status}
            </span>
          )}
        </div>
        {renderActions && (
          <div className="flex items-center gap-2">{renderActions()}</div>
        )}
      </div>
      {backLink && <div>{backLink}</div>}
      {children}
    </div>
  )
}
