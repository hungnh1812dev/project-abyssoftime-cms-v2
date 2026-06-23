import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Badge } from '../badge';

describe('Badge', () => {
  it('renders children with default variant', () => {
    render(<Badge>Default</Badge>);
    expect(screen.getByText('Default')).toBeInTheDocument();
  });

  it('applies draft variant classes', () => {
    render(<Badge variant="draft">Draft</Badge>);
    const el = screen.getByText('Draft');
    expect(el.className).toContain('bg-warning/10');
    expect(el.className).toContain('text-warning');
    expect(el.className).toContain('border-warning/20');
  });

  it('applies published variant classes', () => {
    render(<Badge variant="published">Published</Badge>);
    const el = screen.getByText('Published');
    expect(el.className).toContain('bg-success/10');
    expect(el.className).toContain('text-success');
    expect(el.className).toContain('border-success/20');
  });

  it('applies modified variant classes', () => {
    render(<Badge variant="modified">Modified</Badge>);
    const el = screen.getByText('Modified');
    expect(el.className).toContain('bg-primary-muted');
    expect(el.className).toContain('text-primary');
    expect(el.className).toContain('border-primary/20');
  });
});
