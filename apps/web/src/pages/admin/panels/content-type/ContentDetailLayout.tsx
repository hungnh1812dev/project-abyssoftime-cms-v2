import type { ReactNode } from 'react';
import { StickyActionBar } from '@/pages/admin/layout/StickyActionBar';
import type { BreadcrumbItem } from '@/hooks/useBreadcrumbs';

interface Props {
  title: string;
  status?: string;
  breadcrumbs?: BreadcrumbItem[];
  backLink?: ReactNode;
  metadata?: ReactNode;
  renderActions?: () => ReactNode;
  children: ReactNode;
}

export function ContentDetailLayout({ title, status, breadcrumbs, backLink, metadata, renderActions, children }: Props) {
  return (
    <div className="min-h-full">
      <StickyActionBar title={title} status={status} breadcrumbs={breadcrumbs} renderActions={renderActions} />
      <div className="p-6">
        {backLink && <div className="mb-0.5">{backLink}</div>}
        {metadata && <div className="mb-4">{metadata}</div>}
        {children}
      </div>
    </div>
  );
}
