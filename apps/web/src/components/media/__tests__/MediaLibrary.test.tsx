import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';
import { api } from '@/lib/api';
import { renderWithProviders } from '@/test-utils';
import { MediaLibrary } from '../MediaLibrary';

let mock: MockAdapter;

const mediaResponse = {
  items: [
    {
      ID: 'a1',
      url: 'https://cdn/a1.jpg',
      thumbnailUrl: 'https://cdn/a1.jpg',
      publicId: 'p1',
      fileName: 'a1_abc.jpg',
      fileExt: 'jpg',
      hash: 'abc',
      documentRef: '',
      contentTypeId: '',
      createdAt: '',
    },
    {
      ID: 'a2',
      url: 'https://cdn/a2.jpg',
      thumbnailUrl: 'https://cdn/a2.jpg',
      publicId: 'p2',
      fileName: 'a2_def.jpg',
      fileExt: 'jpg',
      hash: 'def',
      documentRef: '',
      contentTypeId: '',
      createdAt: '',
    },
  ],
  total: 2,
  page: 1,
  limit: 20,
};

beforeEach(() => {
  mock = new MockAdapter(api);
});

afterEach(() => {
  mock.restore();
  vi.clearAllMocks();
});

describe('MediaLibrary', () => {
  it('renders thumbnails from API when open', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getAllByRole('img')).toHaveLength(2);
    });
  });

  it('calls onSelect and onClose when a thumbnail is clicked', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);
    const onSelect = vi.fn();
    const onClose = vi.fn();

    renderWithProviders(<MediaLibrary isOpen onClose={onClose} onSelect={onSelect} />);

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2));
    await userEvent.click(screen.getAllByRole('img')[0]);

    expect(onSelect).toHaveBeenCalledWith(mediaResponse.items[0]);
    expect(onClose).toHaveBeenCalled();
  });

  it('does not render when isOpen is false', () => {
    renderWithProviders(<MediaLibrary isOpen={false} onClose={vi.fn()} onSelect={vi.fn()} />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('shows prev/next pagination buttons', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, { ...mediaResponse, total: 50 });

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument();
    });
  });

  // ---- Delete ----------------------------------------------------------------

  it('renders a delete button for each asset tile', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2));
    expect(screen.getAllByRole('button', { name: 'Delete asset' })).toHaveLength(2);
  });

  it('arms confirm state on first delete click (does not call API)', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);
    let deleteCalled = false;
    mock.onDelete('/api/media/a1').reply(() => {
      deleteCalled = true;
      return [204];
    });

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2));
    const [firstDeleteBtn] = screen.getAllByRole('button', { name: 'Delete asset' });
    await userEvent.click(firstDeleteBtn);

    expect(screen.getByRole('button', { name: 'Confirm delete' })).toBeInTheDocument();
    expect(deleteCalled).toBe(false);
  });

  it('fires DELETE on second click and invalidates the list', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);
    mock.onDelete('/api/media/a1').reply(204);

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2));
    const [firstDeleteBtn] = screen.getAllByRole('button', { name: 'Delete asset' });
    await userEvent.click(firstDeleteBtn);

    const confirmBtn = screen.getByRole('button', { name: 'Confirm delete' });
    await userEvent.click(confirmBtn);

    await waitFor(() => expect(mock.history.delete).toHaveLength(1));
    expect(mock.history.delete[0].url).toBe('/api/media/a1');
  });

  it('disarms confirm state on mouse-leave', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);

    renderWithProviders(<MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />);

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2));
    const [firstDeleteBtn] = screen.getAllByRole('button', { name: 'Delete asset' });
    await userEvent.click(firstDeleteBtn);

    expect(screen.getByRole('button', { name: 'Confirm delete' })).toBeInTheDocument();

    await userEvent.unhover(firstDeleteBtn);

    await waitFor(() => expect(screen.queryByRole('button', { name: 'Confirm delete' })).not.toBeInTheDocument());
  });
});
