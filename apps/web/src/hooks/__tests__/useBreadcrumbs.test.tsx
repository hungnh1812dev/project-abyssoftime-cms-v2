import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { useBreadcrumbs } from '../useBreadcrumbs'

function wrapper(path: string) {
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter initialEntries={[path]}>{children}</MemoryRouter>
  )
}

describe('useBreadcrumbs', () => {
  it('returns Home for /admin', () => {
    const { result } = renderHook(() => useBreadcrumbs(), { wrapper: wrapper('/admin') })
    expect(result.current).toEqual([{ label: 'Home', to: '/admin' }])
  })

  it('returns Content Manager > slug for single-type route', () => {
    const { result } = renderHook(() => useBreadcrumbs(), {
      wrapper: wrapper('/admin/content-type/single-type/homepage'),
    })
    expect(result.current).toEqual([
      { label: 'Home', to: '/admin' },
      { label: 'Content Manager' },
      { label: 'homepage' },
    ])
  })

  it('returns Content Manager > slug for collection-type route', () => {
    const { result } = renderHook(() => useBreadcrumbs(), {
      wrapper: wrapper('/admin/content-type/collection-type/articles'),
    })
    expect(result.current).toEqual([
      { label: 'Home', to: '/admin' },
      { label: 'Content Manager' },
      { label: 'articles' },
    ])
  })

  it('returns Settings > page name for settings route', () => {
    const { result } = renderHook(() => useBreadcrumbs(), {
      wrapper: wrapper('/admin/settings/media'),
    })
    expect(result.current).toEqual([
      { label: 'Home', to: '/admin' },
      { label: 'Settings' },
      { label: 'Media' },
    ])
  })

  it('returns Settings > Users for settings/users', () => {
    const { result } = renderHook(() => useBreadcrumbs(), {
      wrapper: wrapper('/admin/settings/users'),
    })
    expect(result.current).toEqual([
      { label: 'Home', to: '/admin' },
      { label: 'Settings' },
      { label: 'Users' },
    ])
  })
})
