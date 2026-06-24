import type { ReactNode } from 'react';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb } from '@/components/ui/breadcrumb';
import type { BreadcrumbItem } from '@/hooks/useBreadcrumbs';

const STATUS_VARIANT: Record<string, 'draft' | 'published' | 'modified'> = {
  draft: 'draft',
  published: 'published',
  modified: 'modified',
};

interface StickyActionBarProps {
  title: string;
  status?: string;
  breadcrumbs?: BreadcrumbItem[];
  renderActions?: () => ReactNode;
}

export function StickyActionBar({ title, status, breadcrumbs, renderActions }: StickyActionBarProps) {
  return (
    <div className="border-border bg-background/80 sticky top-0 z-30 flex min-h-16 items-center justify-between border-b px-6 py-2 backdrop-blur-sm">
      <div className="flex flex-col justify-center gap-0.5">
        {breadcrumbs && breadcrumbs.length > 0 && <Breadcrumb items={breadcrumbs} />}
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold">{title}</h1>
          {status && (
            <Badge data-testid="status-badge" variant={STATUS_VARIANT[status] ?? 'secondary'} className="capitalize">
              {status}
            </Badge>
          )}
        </div>
      </div>
      {renderActions && <div className="flex items-center gap-2">{renderActions()}</div>}
    </div>
  );
}
