import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from '../Sidebar'
import { SidebarProvider } from '../SidebarContext'

vi.mock('@/hooks/useContentTypes', () => ({
  useContentTypes: () => ({
    data: [
      { ID: '1', Name: 'Homepage', Slug: 'homepage', Kind: 'single' },
      { ID: '2', Name: 'Articles', Slug: 'articles', Kind: 'collection' },
    ],
  }),
}))

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({ role: 'super_admin', token: 'x', userId: '1', loading: false, login: vi.fn(), logout: vi.fn() }),
}))

function renderSidebar(initialPath = '/admin') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <SidebarProvider>
        <Sidebar />
      </SidebarProvider>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  localStorage.clear()
})

describe('Sidebar', () => {
  it('renders brand text', () => {
    renderSidebar()
    expect(screen.getByText('AbyssOfTime CMS')).toBeInTheDocument()
  })

  it('renders Content Manager group with content type items', () => {
    renderSidebar()
    expect(screen.getByText('Content Manager')).toBeInTheDocument()
    expect(screen.getByText('Homepage')).toBeInTheDocument()
    expect(screen.getByText('Articles')).toBeInTheDocument()
  })

  it('renders Single Types and Collection Types sub-groups', () => {
    renderSidebar()
    expect(screen.getByText('Single Types')).toBeInTheDocument()
    expect(screen.getByText('Collection Types')).toBeInTheDocument()
  })

  it('renders Settings group with nav items', () => {
    renderSidebar()
    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('Media Library')).toBeInTheDocument()
    expect(screen.getByText('Users')).toBeInTheDocument()
    expect(screen.getByText('Access Tokens')).toBeInTheDocument()
    expect(screen.getByText('Roles')).toBeInTheDocument()
  })

  it('renders collapse toggle button', () => {
    renderSidebar()
    expect(screen.getByRole('button', { name: /collapse/i })).toBeInTheDocument()
  })

  it('collapses sidebar when toggle is clicked', async () => {
    renderSidebar()
    const sidebar = screen.getByRole('complementary')
    expect(sidebar.className).toContain('w-64')

    await userEvent.click(screen.getByRole('button', { name: /collapse/i }))
    expect(sidebar.className).toContain('w-16')
  })

  it('hides text labels when collapsed', async () => {
    renderSidebar()
    await userEvent.click(screen.getByRole('button', { name: /collapse/i }))
    const brand = screen.queryByText('AbyssOfTime CMS')
    expect(brand).not.toBeVisible()
  })

  it('generates correct links for content types', () => {
    renderSidebar()
    const homepageLink = screen.getByRole('link', { name: 'Homepage' })
    expect(homepageLink).toHaveAttribute('href', '/admin/content-type/single-type/homepage')
    const articlesLink = screen.getByRole('link', { name: 'Articles' })
    expect(articlesLink).toHaveAttribute('href', '/admin/content-type/collection-type/articles')
  })

  it('generates correct links for settings items', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: 'Media Library' })).toHaveAttribute('href', '/admin/settings/media')
    expect(screen.getByRole('link', { name: 'Users' })).toHaveAttribute('href', '/admin/settings/users')
  })
})
