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
  it('renders a Choose file button', () => {
    const mutate = vi.fn()
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate,
      isPending: false,
    } as ReturnType<typeof useUploadMedia>)

    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByRole('button', { name: /choose file/i })).toBeInTheDocument()
  })

  it('shows Uploading… while upload is in flight', () => {
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate: vi.fn(),
      isPending: true,
    } as ReturnType<typeof useUploadMedia>)

    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByRole('button', { name: /uploading/i })).toBeDisabled()
  })

  it('calls upload and shows preview on file select', async () => {
    const user = userEvent.setup()
    let onSuccessCb: ((asset: { url: string }) => void) | undefined

    const mutate = vi.fn((_args, opts) => {
      onSuccessCb = opts?.onSuccess
    })
    vi.mocked(useUploadMedia).mockReturnValue({
      mutate,
      isPending: false,
    } as ReturnType<typeof useUploadMedia>)

    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
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

    // Simulate successful upload
    await act(async () => {
      onSuccessCb?.({ url: 'https://cdn.example.com/photo.jpg' })
    })

    await waitFor(() => {
      expect(screen.getByRole('img', { name: /uploaded media/i })).toHaveAttribute(
        'src',
        'https://cdn.example.com/photo.jpg',
      )
    })
  })
})
