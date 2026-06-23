import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from '../button';

describe('Button', () => {
  it('renders children', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });

  it('applies success variant classes', () => {
    render(<Button variant="success">Publish</Button>);
    const btn = screen.getByRole('button', { name: 'Publish' });
    expect(btn.className).toContain('bg-success');
    expect(btn.className).toContain('text-success-foreground');
  });

  it('renders loading spinner and sets aria-busy when loading', () => {
    render(<Button loading>Save</Button>);
    const btn = screen.getByRole('button');
    expect(btn).toHaveAttribute('aria-busy', 'true');
    expect(btn.querySelector('.animate-spin')).toBeInTheDocument();
  });

  it('hides children and shows loadingText when loading', () => {
    render(
      <Button loading loadingText="Saving...">
        Save
      </Button>,
    );
    const btn = screen.getByRole('button');
    expect(btn).toHaveTextContent('Saving...');
    expect(btn).not.toHaveTextContent('Save');
  });

  it('prevents clicks when loading', async () => {
    const onClick = vi.fn();
    render(
      <Button loading onClick={onClick}>
        Save
      </Button>,
    );
    const btn = screen.getByRole('button');
    await userEvent.click(btn);
    expect(onClick).not.toHaveBeenCalled();
  });

  it('applies default variant hover classes with primary-hover', () => {
    render(<Button variant="default">Action</Button>);
    const btn = screen.getByRole('button', { name: 'Action' });
    expect(btn.className).toContain('hover:bg-primary-hover');
  });

  it('applies lg size with h-10', () => {
    render(<Button size="lg">Large</Button>);
    const btn = screen.getByRole('button', { name: 'Large' });
    expect(btn.className).toContain('h-10');
  });
});
