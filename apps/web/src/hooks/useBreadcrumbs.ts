import { useMemo } from 'react'
import { useLocation } from 'react-router-dom'

export interface BreadcrumbItem {
  label: string
  to?: string
}

function slugToTitle(slug: string): string {
  return slug.replace(/[-_]/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

const SETTINGS_LABELS: Record<string, string> = {
  media: 'Media',
  users: 'Users',
  'access-tokens': 'Access Tokens',
  roles: 'Roles',
}

export function useBreadcrumbs(): BreadcrumbItem[] {
  const { pathname } = useLocation()

  return useMemo(() => {
    const crumbs: BreadcrumbItem[] = [{ label: 'Home', to: '/admin' }]

    if (pathname === '/admin' || pathname === '/admin/') {
      return crumbs
    }

    const rest = pathname.replace(/^\/admin\/?/, '')

    if (rest.startsWith('content-type/')) {
      crumbs.push({ label: 'Content Manager' })
      const parts = rest.split('/')
      const slug = parts[2]
      if (slug) {
        crumbs.push({ label: slugToTitle(slug) })
      }
      return crumbs
    }

    if (rest.startsWith('settings/')) {
      crumbs.push({ label: 'Settings' })
      const page = rest.replace('settings/', '')
      if (page) {
        crumbs.push({ label: SETTINGS_LABELS[page] ?? page })
      }
      return crumbs
    }

    return crumbs
  }, [pathname])
}
