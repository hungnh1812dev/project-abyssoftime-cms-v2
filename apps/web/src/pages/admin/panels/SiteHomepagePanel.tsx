import { ContentTypeLayout, type ContentTypeLayoutProps } from '@/components/content-type/ContentTypeLayout'

export function SiteHomepagePanel({ children, ...props }: ContentTypeLayoutProps) {
  return (
    <ContentTypeLayout {...props}>
      <p className="text-sm text-muted-foreground">
        Configure your site&apos;s homepage content and metadata.
      </p>
      {children}
    </ContentTypeLayout>
  )
}
