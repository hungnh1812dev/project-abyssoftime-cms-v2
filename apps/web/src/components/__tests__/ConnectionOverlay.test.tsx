import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConnectionOverlay } from '@/components/ConnectionOverlay';

describe('ConnectionOverlay', () => {
  it('renders spinner and text when visible', () => {
    render(<ConnectionOverlay visible={true} />);

    expect(screen.getByText('Connecting to service...')).toBeInTheDocument();
    expect(
      screen.getByText('The server may be starting up. This can take up to 30 seconds.'),
    ).toBeInTheDocument();

    const overlay = screen.getByRole('alert');
    expect(overlay).not.toHaveClass('opacity-0');
    expect(overlay).not.toHaveClass('pointer-events-none');
  });

  it('is hidden with opacity-0 and pointer-events-none when not visible', () => {
    render(<ConnectionOverlay visible={false} />);

    const overlay = screen.getByRole('alert');
    expect(overlay).toHaveClass('opacity-0');
    expect(overlay).toHaveClass('pointer-events-none');
  });

  it('has correct accessibility attributes', () => {
    render(<ConnectionOverlay visible={true} />);

    const overlay = screen.getByRole('alert');
    expect(overlay).toHaveAttribute('aria-live', 'assertive');
    expect(overlay).toHaveAttribute('aria-busy', 'true');
  });
});
