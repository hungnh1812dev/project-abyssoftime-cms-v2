import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FormProvider } from '../../FormProvider'
import { FormField } from '../../FormField'

// CodeMirror cannot run in jsdom — mock it with a controlled textarea
vi.mock('@uiw/react-codemirror', () => ({
  default: ({
    value,
    onChange,
  }: {
    value: string
    onChange: (val: string) => void
  }) => (
    <textarea
      data-testid="codemirror-mock"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}))
vi.mock('@codemirror/lang-json', () => ({ json: () => [] }))

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>
}

// Import after mocks are registered
const { JsonInput } = await import('../JsonInput')

describe('JsonInput', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the CodeMirror editor', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="metadata">
            <JsonInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByTestId('codemirror-mock')).toBeInTheDocument()
  })

  it('wraps the editor in a min-h-[15em] container', () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="metadata">
            <JsonInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    const wrapper = screen.getByTestId('json-editor-wrapper')
    expect(wrapper).toHaveClass('min-h-[15em]')
  })

  it('parses valid JSON and submits as an object (not a string)', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="metadata">
            <JsonInput />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )
    // Use fireEvent.change to avoid userEvent brace-escaping for JSON
    fireEvent.change(screen.getByTestId('codemirror-mock'), {
      target: { value: '{"key":"val"}' },
    })
    await user.click(screen.getByRole('button', { name: /submit/i }))
    await waitFor(() => {
      expect(mutationFn.mock.calls[0][0]).toEqual({ metadata: { key: 'val' } })
    })
  })

  it('shows an inline error for invalid JSON', async () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="metadata">
            <JsonInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    fireEvent.change(screen.getByTestId('codemirror-mock'), {
      target: { value: '{bad json}' },
    })
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('blocks form submission when JSON is invalid', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="metadata">
            <JsonInput />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )
    fireEvent.change(screen.getByTestId('codemirror-mock'), {
      target: { value: '{bad json}' },
    })
    await user.click(screen.getByRole('button', { name: /submit/i }))
    await waitFor(() => {
      expect(mutationFn).not.toHaveBeenCalled()
    })
  })
})
