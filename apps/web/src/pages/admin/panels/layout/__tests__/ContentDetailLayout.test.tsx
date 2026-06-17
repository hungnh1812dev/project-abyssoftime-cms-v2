import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ContentDetailLayout } from '../ContentDetailLayout'

describe('ContentDetailLayout', () => {
  it('renders title', () => {
    render(<ContentDetailLayout title="Blog Posts"><p>form</p></ContentDetailLayout>)
    expect(screen.getByRole('heading', { name: 'Blog Posts' })).toBeInTheDocument()
  })

  it('renders status badge when provided', () => {
    render(<ContentDetailLayout title="Posts" status="draft"><span /></ContentDetailLayout>)
    expect(screen.getByTestId('status-badge')).toHaveTextContent('draft')
  })

  it('does not render status badge when status is omitted', () => {
    render(<ContentDetailLayout title="Posts"><span /></ContentDetailLayout>)
    expect(screen.queryByTestId('status-badge')).not.toBeInTheDocument()
  })

  it('renders backLink when provided', () => {
    render(
      <ContentDetailLayout title="Post" backLink={<a href="/list">← Back</a>}>
        <span />
      </ContentDetailLayout>,
    )
    expect(screen.getByRole('link', { name: /back/i })).toBeInTheDocument()
  })

  it('does not render backLink area when omitted', () => {
    render(<ContentDetailLayout title="Post"><span /></ContentDetailLayout>)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('renders renderActions output in the header row', () => {
    render(
      <ContentDetailLayout title="Post" renderActions={() => <button>Publish</button>}>
        <span />
      </ContentDetailLayout>,
    )
    expect(screen.getByRole('button', { name: /publish/i })).toBeInTheDocument()
  })

  it('renders children below the header', () => {
    render(
      <ContentDetailLayout title="Post">
        <p data-testid="child-content">form here</p>
      </ContentDetailLayout>,
    )
    expect(screen.getByTestId('child-content')).toBeInTheDocument()
  })
})
