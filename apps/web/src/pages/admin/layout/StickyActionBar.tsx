import type { ReactNode } from 'react'
import { Badge } from '@/components/ui/badge'

const STATUS_VARIANT: Record<string, 'draft' | 'published' | 'modified'> = {
  draft: 'draft',
  published: 'published',
  modified: 'modified',
}

interface StickyActionBarProps {
  title: string
  status?: string
  renderActions?: () => ReactNode
}

export function StickyActionBar({ title, status, renderActions }: StickyActionBarProps) {
  return (
    <div className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 px-6 backdrop-blur-sm">
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold">{title}</h1>
        {status && (
          <Badge
            data-testid="status-badge"
            variant={STATUS_VARIANT[status] ?? 'secondary'}
            className="capitalize"
          >
            {status}
          </Badge>
        )}
      </div>
      {renderActions && (
        <div className="flex items-center gap-2">{renderActions()}</div>
      )}
    </div>
  )
}
