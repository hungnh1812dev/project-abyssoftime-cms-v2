import type { ReactNode } from 'react';
import { StickyActionBar } from '@/pages/admin/layout/StickyActionBar';

export interface ContentTypeLayoutProps {
  title: string;
  status?: string;
  renderHeader?: (defaultHeader: ReactNode) => ReactNode;
  renderActions?: () => ReactNode;
  children: ReactNode;
}

export function ContentTypeLayout({ title, status, renderHeader, renderActions, children }: ContentTypeLayoutProps) {
  if (renderHeader) {
    const defaultHeader = (
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold">{title}</h1>
        {status && (
          <span data-testid="status-badge" className="text-muted-foreground text-sm capitalize">
            {status}
          </span>
        )}
      </div>
    );

    return (
      <div className="min-h-full">
        <div className="border-border bg-background/80 sticky top-0 z-30 flex h-16 items-center justify-between border-b px-6 backdrop-blur-sm">{renderHeader(defaultHeader)}</div>
        <div className="p-6">{children}</div>
      </div>
    );
  }

  return (
    <div className="min-h-full">
      <StickyActionBar title={title} status={status} renderActions={renderActions} />
      <div className="p-6">{children}</div>
    </div>
  );
}
