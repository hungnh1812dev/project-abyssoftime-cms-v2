import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { FormProvider } from '../FormProvider'
import { FormField } from '../FormField'
import { useCmsFormState } from '../FormStateContext'

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>
}

describe('FormProvider + FormField', () => {
  it('serializes dot-notation field names into nested JSON on submit', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="a.b">
            <input aria-label="nested-input" />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )

    await user.type(screen.getByLabelText('nested-input'), 'hello')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      expect(mutationFn).toHaveBeenCalledTimes(1)
      expect(mutationFn.mock.calls[0][0]).toEqual({ a: { b: 'hello' } })
    })
  })

  it('renders FormField child with register props injected', async () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="title">
            <input aria-label="title-input" />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )

    expect(screen.getByLabelText('title-input')).toBeInTheDocument()
  })

  it('does not use React.Children.map in FormProvider or FormField source', () => {
    const dir = import.meta.dirname
    const formProviderSrc = readFileSync(resolve(dir, '../FormProvider.tsx'), 'utf-8')
    const formFieldSrc = readFileSync(resolve(dir, '../FormField.tsx'), 'utf-8')
    expect(formProviderSrc).not.toContain('Children.map')
    expect(formFieldSrc).not.toContain('Children.map')
  })

  it('exposes loading, submitting and isDirty state via context', async () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    function StateConsumer() {
      const { loading, submitting, isDirty } = useCmsFormState()
      return (
        <span data-testid="state">
          {loading ? 'loading' : submitting ? 'submitting' : isDirty ? 'dirty' : 'idle'}
        </span>
      )
    }

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <StateConsumer />
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )

    expect(screen.getByTestId('state')).toHaveTextContent('idle')
  })

  it('isDirty is false on initial load and true after editing a field', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)

    function DirtyConsumer() {
      const { isDirty } = useCmsFormState()
      return <span data-testid="dirty">{isDirty ? 'dirty' : 'clean'}</span>
    }

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <DirtyConsumer />
          <FormField name="title">
            <input aria-label="title" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )

    expect(screen.getByTestId('dirty')).toHaveTextContent('clean')
    await user.type(screen.getByLabelText('title'), 'x')
    expect(screen.getByTestId('dirty')).toHaveTextContent('dirty')
  })

  it('fires success toast and resets form to clean after successful save', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue({ title: 'saved-value' })

    function DirtyConsumer() {
      const { isDirty } = useCmsFormState()
      return <span data-testid="dirty">{isDirty ? 'dirty' : 'clean'}</span>
    }

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <DirtyConsumer />
          <FormField name="title">
            <input aria-label="title" />
          </FormField>
          <button type="submit">Save</button>
        </FormProvider>
      </Wrapper>,
    )

    await user.type(screen.getByLabelText('title'), 'hello')
    expect(screen.getByTestId('dirty')).toHaveTextContent('dirty')

    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Saved')
      expect(screen.getByTestId('dirty')).toHaveTextContent('clean')
    })
  })

  it('fires error toast and preserves edited values on failed save', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockRejectedValue(
      Object.assign(new Error('Bad request'), {
        response: { data: { error: 'Validation failed' } },
      }),
    )

    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="title">
            <input aria-label="title" />
          </FormField>
          <button type="submit">Save</button>
        </FormProvider>
      </Wrapper>,
    )

    await user.type(screen.getByLabelText('title'), 'bad value')
    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Validation failed')
    })
    expect(screen.getByLabelText('title')).toHaveValue('bad value')
  })
})
