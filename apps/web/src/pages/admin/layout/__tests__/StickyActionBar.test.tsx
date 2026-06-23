import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StickyActionBar } from '../StickyActionBar';

describe('StickyActionBar', () => {
  it('renders title', () => {
    render(<StickyActionBar title="Homepage" />);
    expect(screen.getByText('Homepage')).toBeInTheDocument();
  });

  it('renders status badge when provided', () => {
    render(<StickyActionBar title="Post" status="draft" />);
    const badge = screen.getByTestId('status-badge');
    expect(badge).toHaveTextContent('draft');
  });

  it('does not render status badge when omitted', () => {
    render(<StickyActionBar title="Post" />);
    expect(screen.queryByTestId('status-badge')).not.toBeInTheDocument();
  });

  it('renders action buttons from renderActions', () => {
    render(<StickyActionBar title="Post" renderActions={() => <button>Save</button>} />);
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
  });

  it('has sticky positioning classes', () => {
    const { container } = render(<StickyActionBar title="Test" />);
    const bar = container.firstElementChild as HTMLElement;
    expect(bar.className).toContain('sticky');
    expect(bar.className).toContain('top-0');
  });

  it('has glassmorphism background', () => {
    const { container } = render(<StickyActionBar title="Test" />);
    const bar = container.firstElementChild as HTMLElement;
    expect(bar.className).toContain('backdrop-blur');
  });
});
