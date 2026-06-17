import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FormProvider } from '../../FormProvider'
import { FormField } from '../../FormField'
import { MediaInput } from '../MediaInput'

// Mock the upload hook so tests don't hit the network
vi.mock('@/hooks/useMedia', () => ({
  useUploadMedia: vi.fn(),
}))

import { useUploadMedia } from '@/hooks/useMedia'

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('MediaInput', () => {
  it('renders a clickable upload zone with placeholder text', () => {
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as ReturnType<typeof useUploadMedia>)

    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByTestId('media-upload-zone')).toBeInTheDocument()
    expect(screen.getByText(/click to upload/i)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /choose file/i })).not.toBeInTheDocument()
  })

  it('shows a spinner overlay while uploading', () => {
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate: vi.fn(),
      isPending: true,
    } as ReturnType<typeof useUploadMedia>)

    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByTestId('upload-spinner')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /uploading/i })).not.toBeInTheDocument()
  })

  it('calls upload and shows the image in the zone after success', async () => {
    const user = userEvent.setup()
    let onSuccessCb: ((asset: { url: string; thumbnailUrl: string }) => void) | undefined

    const mutate = vi.fn((_args, opts) => {
      onSuccessCb = opts?.onSuccess
    })
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate,
      isPending: false,
    } as ReturnType<typeof useUploadMedia>)

    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput documentRef="doc-1" contentTypeId="ct-1" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )

    const file = new File(['fake'], 'photo.jpg', { type: 'image/jpeg' })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(fileInput, file)

    expect(mutate).toHaveBeenCalledWith(
      expect.objectContaining({ file, documentRef: 'doc-1', contentTypeId: 'ct-1' }),
      expect.any(Object),
    )

    await act(async () => {
      onSuccessCb?.({ url: 'https://cdn.example.com/photo.jpg', thumbnailUrl: 'https://cdn.example.com/photo.jpg' })
    })

    await waitFor(() => {
      expect(screen.getByRole('img', { name: /media preview/i })).toHaveAttribute(
        'src',
        'https://cdn.example.com/photo.jpg',
      )
    })
  })

  it('displays thumbnailUrl in the zone when it differs from url', async () => {
    const user = userEvent.setup()
    let onSuccessCb: ((asset: { url: string; thumbnailUrl: string }) => void) | undefined

    const mutate = vi.fn((_args, opts) => {
      onSuccessCb = opts?.onSuccess
    })
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate,
      isPending: false,
    } as ReturnType<typeof useUploadMedia>)

    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput documentRef="doc-1" contentTypeId="ct-1" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )

    const file = new File(['fake'], 'photo.jpg', { type: 'image/jpeg' })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(fileInput, file)

    await act(async () => {
      onSuccessCb?.({
        url: 'https://cdn.example.com/photo.jpg',
        thumbnailUrl: 'https://cdn.example.com/thumb_photo.jpg',
      })
    })

    await waitFor(() => {
      expect(screen.getByRole('img', { name: /media preview/i })).toHaveAttribute(
        'src',
        'https://cdn.example.com/thumb_photo.jpg',
      )
    })
  })
})
