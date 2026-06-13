import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FormProvider } from '../../FormProvider'
import { FormField } from '../../FormField'

// CKEditor cannot run in jsdom — mock it with a controlled textarea that
// mirrors the real onChange(event, editor) API.
vi.mock('@ckeditor/ckeditor5-react', () => ({
  CKEditor: ({
    data,
    onChange,
  }: {
    data: string
    onChange: (event: null, editor: { getData: () => string }) => void
  }) => (
    <textarea
      data-testid="ckeditor-mock"
      defaultValue={data}
      onChange={(e) => onChange(null, { getData: () => e.target.value })}
    />
  ),
}))

vi.mock('@ckeditor/ckeditor5-build-classic', () => ({ default: class MockEditor {} }))

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>
}

// Import after mocks are registered
const { RichTextInput } = await import('../RichTextInput')

describe('RichTextInput', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the CKEditor', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="body">
            <RichTextInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByTestId('ckeditor-mock')).toBeInTheDocument()
  })

  it('updates form value when editor content changes', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="body">
            <RichTextInput />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )

    const editor = screen.getByTestId('ckeditor-mock')
    await user.clear(editor)
    await user.type(editor, '<p>Hello</p>')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      expect(mutationFn.mock.calls[0][0]).toEqual({ body: '<p>Hello</p>' })
    })
  })

  it('stores HTML string (not DOM nodes) in form state', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="content">
            <RichTextInput />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )

    await user.clear(screen.getByTestId('ckeditor-mock'))
    await user.type(screen.getByTestId('ckeditor-mock'), '<h1>Title</h1>')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      const submitted = mutationFn.mock.calls[0][0] as Record<string, unknown>
      expect(typeof submitted.content).toBe('string')
    })
  })
})
