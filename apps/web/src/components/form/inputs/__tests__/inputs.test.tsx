import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FormProvider } from '../../FormProvider'
import { FormField } from '../../FormField'
import { TextInput } from '../TextInput'
import { NumberInput } from '../NumberInput'
import { BooleanInput } from '../BooleanInput'

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>
}

// ──────────────────────────────────────────────
// TextInput
// ──────────────────────────────────────────────

describe('TextInput', () => {
  it('renders a text input', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="title">
            <TextInput aria-label="title" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByRole('textbox', { name: 'title' })).toBeInTheDocument()
  })

  it('accepts typed value and submits it', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="title">
            <TextInput aria-label="title" />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )
    await user.type(screen.getByRole('textbox', { name: 'title' }), 'Hello world')
    await user.click(screen.getByRole('button', { name: /submit/i }))
    await waitFor(() => {
      expect(mutationFn.mock.calls[0][0]).toEqual({ title: 'Hello world' })
    })
  })

  it('renders a textarea when multiline is true', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="body">
            <TextInput multiline aria-label="body" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByRole('textbox', { name: 'body' })).toBeInTheDocument()
    // Textarea has no type attribute — input does; distinguish via tagName
    const el = screen.getByRole('textbox', { name: 'body' })
    expect(el.tagName.toLowerCase()).toBe('textarea')
  })
})

// ──────────────────────────────────────────────
// NumberInput
// ──────────────────────────────────────────────

describe('NumberInput', () => {
  it('renders a number input', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="count">
            <NumberInput aria-label="count" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    expect(screen.getByRole('spinbutton', { name: 'count' })).toBeInTheDocument()
  })

  it('passes step, min, max to the underlying input', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="price">
            <NumberInput aria-label="price" step={0.01} min={0} max={9999} />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    const el = screen.getByRole('spinbutton', { name: 'price' })
    expect(el).toHaveAttribute('step', '0.01')
    expect(el).toHaveAttribute('min', '0')
    expect(el).toHaveAttribute('max', '9999')
  })
})

// ──────────────────────────────────────────────
// BooleanInput
// ──────────────────────────────────────────────

describe('BooleanInput', () => {
  it('renders a switch / checkbox', () => {
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="active">
            <BooleanInput aria-label="active" />
          </FormField>
        </FormProvider>
      </Wrapper>,
    )
    // @base-ui Switch renders with role="switch"
    expect(screen.getByRole('switch', { name: 'active' })).toBeInTheDocument()
  })

  it('toggles and submits boolean value', async () => {
    const user = userEvent.setup()
    const mutationFn = vi.fn().mockResolvedValue(undefined)
    render(
      <Wrapper>
        <FormProvider mutationFn={mutationFn}>
          <FormField name="active">
            <BooleanInput aria-label="active" />
          </FormField>
          <button type="submit">Submit</button>
        </FormProvider>
      </Wrapper>,
    )
    await user.click(screen.getByRole('switch', { name: 'active' }))
    await user.click(screen.getByRole('button', { name: /submit/i }))
    await waitFor(() => {
      expect(mutationFn.mock.calls[0][0]).toEqual({ active: true })
    })
  })
})
