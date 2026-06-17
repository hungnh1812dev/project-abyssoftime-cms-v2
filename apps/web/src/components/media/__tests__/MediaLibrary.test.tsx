import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { MediaLibrary } from '../MediaLibrary'

let mock: MockAdapter

const mediaResponse = {
  items: [
    { ID: 'a1', url: 'https://cdn/a1.jpg', thumbnailUrl: 'https://cdn/a1.jpg', publicId: 'p1', fileName: 'a1_abc.jpg', fileExt: 'jpg', hash: 'abc', documentRef: '', contentTypeId: '', createdAt: '' },
    { ID: 'a2', url: 'https://cdn/a2.jpg', thumbnailUrl: 'https://cdn/a2.jpg', publicId: 'p2', fileName: 'a2_def.jpg', fileExt: 'jpg', hash: 'def', documentRef: '', contentTypeId: '', createdAt: '' },
  ],
  total: 2,
  page: 1,
  limit: 20,
}

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('MediaLibrary', () => {
  it('renders thumbnails from API when open', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse)

    renderWithProviders(
      <MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getAllByRole('img')).toHaveLength(2)
    })
  })

  it('calls onSelect and onClose when a thumbnail is clicked', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse)
    const onSelect = vi.fn()
    const onClose = vi.fn()

    renderWithProviders(
      <MediaLibrary isOpen onClose={onClose} onSelect={onSelect} />,
    )

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(2))
    await userEvent.click(screen.getAllByRole('img')[0])

    expect(onSelect).toHaveBeenCalledWith(mediaResponse.items[0])
    expect(onClose).toHaveBeenCalled()
  })

  it('does not render when isOpen is false', () => {
    renderWithProviders(
      <MediaLibrary isOpen={false} onClose={vi.fn()} onSelect={vi.fn()} />,
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows prev/next pagination buttons', async () => {
    mock.onGet('/api/media?page=1&limit=20').reply(200, { ...mediaResponse, total: 50 })

    renderWithProviders(
      <MediaLibrary isOpen onClose={vi.fn()} onSelect={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument()
    })
  })
})
