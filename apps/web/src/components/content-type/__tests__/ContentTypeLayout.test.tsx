import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ContentTypeLayout } from '../ContentTypeLayout'

describe('ContentTypeLayout', () => {
  it('renders title and children', () => {
    render(
      <ContentTypeLayout title="Homepage">
        <p>form goes here</p>
      </ContentTypeLayout>,
    )
    expect(screen.getByText('Homepage')).toBeInTheDocument()
    expect(screen.getByText('form goes here')).toBeInTheDocument()
  })

  it('renders status badge when status is provided', () => {
    render(<ContentTypeLayout title="Post" status="draft"><span /></ContentTypeLayout>)
    expect(screen.getByText('draft')).toBeInTheDocument()
  })

  it('does not render a status badge when status is omitted', () => {
    render(<ContentTypeLayout title="Post"><span /></ContentTypeLayout>)
    expect(screen.queryByTestId('status-badge')).not.toBeInTheDocument()
  })

  it('renders renderActions output to the right of the default header', () => {
    render(
      <ContentTypeLayout title="Blog" renderActions={() => <button>Publish</button>}>
        <span />
      </ContentTypeLayout>,
    )
    expect(screen.getByRole('button', { name: /publish/i })).toBeInTheDocument()
  })

  it('renderHeader replaces the entire header row', () => {
    render(
      <ContentTypeLayout
        title="Blog"
        status="published"
        renderHeader={() => <div data-testid="custom-header">Custom</div>}
      >
        <span />
      </ContentTypeLayout>,
    )
    expect(screen.getByTestId('custom-header')).toBeInTheDocument()
    expect(screen.queryByText('Blog')).not.toBeInTheDocument()
    expect(screen.queryByText('published')).not.toBeInTheDocument()
  })

  it('renderHeader receives the default header node as its argument', () => {
    let capturedDefault: React.ReactNode = null
    render(
      <ContentTypeLayout
        title="MyType"
        renderHeader={(def) => {
          capturedDefault = def
          return <div data-testid="wrapper">{def}</div>
        }}
      >
        <span />
      </ContentTypeLayout>,
    )
    expect(capturedDefault).not.toBeNull()
    expect(screen.getByTestId('wrapper')).toBeInTheDocument()
  })
})
