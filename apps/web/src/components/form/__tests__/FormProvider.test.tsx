import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { FormProvider } from '../FormProvider'
import { FormField } from '../FormField'

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
      // TanStack Query v5 passes a context object as the second arg — check only the form data
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

  it('exposes loading and submitting state via context', async () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    const { useCmsFormState } = await import('../FormStateContext')

    function StateConsumer() {
      const { loading, submitting } = useCmsFormState()
      return (
        <span data-testid="state">{loading ? 'loading' : submitting ? 'submitting' : 'idle'}</span>
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
})
