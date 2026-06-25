import { describe, it, expect, vi } from 'vitest';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test-utils';
import { ContentTypeBuilder } from '../ContentTypeBuilder';
import type { FieldDefinition } from '@/types/cms';

const noop = () => Promise.resolve();

describe('ContentTypeBuilder — primitives', () => {
  it('renders a text input for type=text', () => {
    const schema: FieldDefinition[] = [{ name: 'title', type: 'text' }];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    expect(screen.getByLabelText('title')).toBeInTheDocument();
  });

  it('renders a number input for type=number', () => {
    const schema: FieldDefinition[] = [{ name: 'price', type: 'number' }];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    expect(screen.getByLabelText('price')).toBeInTheDocument();
  });

  it('renders a boolean switch for type=boolean', () => {
    const schema: FieldDefinition[] = [{ name: 'active', type: 'boolean' }];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    expect(screen.getByRole('switch', { name: 'active' })).toBeInTheDocument();
  });
});

describe('ContentTypeBuilder — width', () => {
  it('renders fields in a 6-column grid with correct col-span classes', () => {
    const schema: FieldDefinition[] = [
      { name: 'fullWidth', type: 'text' },
      { name: 'half', type: 'text', width: '50%' },
      { name: 'third', type: 'text', width: '1/3' },
    ];
    const { container } = renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const grid = container.querySelector('.md\\:grid-cols-6');
    expect(grid).toBeInTheDocument();
    expect(screen.getByLabelText('fullWidth').closest('.md\\:col-span-6')).toBeInTheDocument();
    expect(screen.getByLabelText('half').closest('.md\\:col-span-3')).toBeInTheDocument();
    expect(screen.getByLabelText('third').closest('.md\\:col-span-2')).toBeInTheDocument();
  });
});

describe('ContentTypeBuilder — component', () => {
  it('renders a fieldset with legend for type=component', () => {
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    expect(screen.getByRole('group', { name: 'banner' })).toBeInTheDocument();
  });

  it('uses dot-notation field names inside components', async () => {
    const onSubmit = vi.fn();
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={onSubmit} />);
    const input = screen.getByLabelText('title');
    await userEvent.clear(input);
    await userEvent.type(input, 'Hello');
    const btn = screen.getByRole('button', { name: /save/i });
    await userEvent.click(btn);
    // TanStack Query v5 calls mutationFn(variables, context); check first arg only
    const firstCallData = onSubmit.mock.calls[0][0] as Record<string, unknown>;
    expect(firstCallData).toMatchObject({ banner: { title: 'Hello' } });
  });
});

describe('ContentTypeBuilder — collapsible components', () => {
  it('top-level component (depth=0) is expanded by default', () => {
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const group = screen.getByRole('group', { name: 'banner' });
    expect(within(group).getByLabelText('title')).toBeInTheDocument();
    expect(within(group).getByRole('button', { name: /banner/i })).toHaveAttribute('aria-expanded', 'true');
  });

  it('nested component (depth>=1) is collapsed by default', () => {
    const schema: FieldDefinition[] = [
      {
        name: 'section',
        type: 'component',
        fields: [
          {
            name: 'inner',
            type: 'component',
            fields: [{ name: 'subtitle', type: 'text' }],
          },
        ],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const innerGroup = screen.getByRole('group', { name: 'inner' });
    expect(innerGroup).toBeInTheDocument();
    expect(within(innerGroup).queryByLabelText('subtitle')).not.toBeInTheDocument();
    expect(within(innerGroup).getByRole('button', { name: /inner/i })).toHaveAttribute('aria-expanded', 'false');
  });

  it('clicking header toggles expand/collapse', async () => {
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const group = screen.getByRole('group', { name: 'banner' });
    const toggle = within(group).getByRole('button', { name: /banner/i });

    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    expect(within(group).getByLabelText('title')).toBeInTheDocument();

    await userEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(within(group).queryByLabelText('title')).not.toBeInTheDocument();

    await userEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    expect(within(group).getByLabelText('title')).toBeInTheDocument();
  });

  it('shows hint text from first text field value', async () => {
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }, { name: 'count', type: 'number' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const input = screen.getByLabelText('title');
    await userEvent.type(input, 'Hello World');

    const group = screen.getByRole('group', { name: 'banner' });
    expect(within(group).getByText(/Hello World/)).toBeInTheDocument();
  });

  it('shows no hint when component has no text fields', () => {
    const schema: FieldDefinition[] = [
      {
        name: 'stats',
        type: 'component',
        fields: [{ name: 'count', type: 'number' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const group = screen.getByRole('group', { name: 'stats' });
    expect(within(group).queryByText('—')).not.toBeInTheDocument();
  });
});
