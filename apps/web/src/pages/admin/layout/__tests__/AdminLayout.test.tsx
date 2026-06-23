import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';
import { api, setAccessToken, getAccessToken } from '@/lib/api';
import { renderWithProviders } from '@/test-utils';
import { AuthProvider } from '@/context/AuthContext';
import { SidebarShell, SidebarProvider } from '@/components/sidebar';
import { TopBar } from '@/pages/admin/layout/TopBar';
import type { ContentType } from '@/types/cms';

function makeToken(payload: Record<string, unknown>) {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.fakesig`;
}

const ADMIN_TOKEN = makeToken({ userId: 'u1', role: 'admin', exp: 9999999999 });
const SUPER_ADMIN_TOKEN = makeToken({ userId: 'u1', role: 'super_admin', exp: 9999999999 });

const contentTypes: ContentType[] = [
  { ID: '1', DocumentID: 'd1', Name: 'Blog', Slug: 'blog', Kind: 'collection', CreatedAt: '', UpdatedAt: '' },
  { ID: '2', DocumentID: 'd2', Name: 'About', Slug: 'about', Kind: 'single', CreatedAt: '', UpdatedAt: '' },
];

let mock: MockAdapter;

beforeEach(() => {
  mock = new MockAdapter(api);
  setAccessToken(null);
});

afterEach(() => {
  mock.restore();
  vi.clearAllMocks();
});

function renderSidebar(token = SUPER_ADMIN_TOKEN) {
  mock.onPost('/auth/refresh').reply(200, { accessToken: token });
  return renderWithProviders(
    <AuthProvider>
      <SidebarProvider>
        <SidebarShell />
      </SidebarProvider>
    </AuthProvider>,
    { initialEntries: ['/admin'] },
  );
}

describe('Sidebar', () => {
  it('renders content type names fetched from the API', async () => {
    mock.onGet('/api/content-types').reply(200, contentTypes);
    renderSidebar();
    await waitFor(() => expect(screen.getByText('Blog')).toBeInTheDocument());
    expect(screen.getByText('About')).toBeInTheDocument();
  });

  it('renders nav links pointing to new content-type routes by kind', async () => {
    mock.onGet('/api/content-types').reply(200, contentTypes);
    renderSidebar();
    await waitFor(() => expect(screen.getByRole('link', { name: 'Blog' })).toBeInTheDocument());
    expect(screen.getByRole('link', { name: 'Blog' })).toHaveAttribute('href', '/admin/content-type/collection-type/blog');
    expect(screen.getByRole('link', { name: 'About' })).toHaveAttribute('href', '/admin/content-type/single-type/about');
  });

  it('renders no content-type links when no content types exist', async () => {
    mock.onGet('/api/content-types').reply(200, []);
    renderSidebar();
    await waitFor(() => expect(screen.getByRole('link', { name: /media library/i })).toBeInTheDocument());
    expect(screen.queryByRole('link', { name: 'Blog' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'About' })).toBeNull();
  });

  it('groups content types into Single Types and Collection Types sections', async () => {
    mock.onGet('/api/content-types').reply(200, contentTypes);
    renderSidebar();

    await waitFor(() => expect(screen.getByText('Blog')).toBeInTheDocument());

    const singleHeading = screen.getByText('Single Types');
    const collectionHeading = screen.getByText('Collection Types');
    const aboutLink = screen.getByRole('link', { name: 'About' });
    const blogLink = screen.getByRole('link', { name: 'Blog' });

    expect(singleHeading.compareDocumentPosition(aboutLink) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(aboutLink.compareDocumentPosition(collectionHeading) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(collectionHeading.compareDocumentPosition(blogLink) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('omits a section heading when no content type of that kind exists', async () => {
    mock.onGet('/api/content-types').reply(200, [contentTypes[0]]); // collection only
    renderSidebar();

    await waitFor(() => expect(screen.getByText('Blog')).toBeInTheDocument());
    expect(screen.queryByText('Single Types')).not.toBeInTheDocument();
    expect(screen.getByText('Collection Types')).toBeInTheDocument();
  });

  it('renders a Settings section with a Media Library link to /admin/settings/media', async () => {
    mock.onGet('/api/content-types').reply(200, []);
    renderSidebar();

    await waitFor(() => expect(screen.getByRole('link', { name: /media library/i })).toBeInTheDocument());
    expect(screen.getByRole('link', { name: /media library/i })).toHaveAttribute('href', '/admin/settings/media');
  });

  it('renders a Logout button', async () => {
    mock.onGet('/api/content-types').reply(200, []);
    renderSidebar();
    await waitFor(() => expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument());
  });

  it('clears the access token when Logout is clicked', async () => {
    mock.onGet('/api/content-types').reply(200, []);
    mock.onPost('/auth/logout').reply(200);
    const user = userEvent.setup();
    renderSidebar();
    await waitFor(() => expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /logout/i }));
    expect(getAccessToken()).toBeNull();
  });
});

describe('TopBar', () => {
  it('renders breadcrumbs', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN });
    renderWithProviders(
      <AuthProvider>
        <SidebarProvider>
          <TopBar />
        </SidebarProvider>
      </AuthProvider>,
      { initialEntries: ['/admin/settings/media'] },
    );
    await waitFor(() => expect(screen.getByText('Home')).toBeInTheDocument());
    expect(screen.getByText('Settings')).toBeInTheDocument();
    expect(screen.getByText('Media')).toBeInTheDocument();
  });
});
