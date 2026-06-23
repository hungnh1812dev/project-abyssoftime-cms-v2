import { describe, it, expect, vi } from 'vitest';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test-utils';
import { ContentTypeBuilder } from '@/pages/admin/panels/content-type/ContentTypeBuilder';
import type { FieldDefinition } from '@/types/cms';

const noop = () => Promise.resolve();

describe('RepeatableComponentField', () => {
  const repeatableSchema: FieldDefinition[] = [
    {
      name: 'skills',
      type: 'component',
      repeatable: true,
      fields: [
        { name: 'category', type: 'text' },
        { name: 'skill', type: 'text' },
      ],
    },
  ];

  it('renders an add button when empty', () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    expect(screen.getByRole('button', { name: /add entry/i })).toBeInTheDocument();
  });

  it('adds an entry when add button is clicked', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);

    expect(screen.getByText('#1')).toBeInTheDocument();
    expect(screen.getByLabelText('category')).toBeInTheDocument();
    expect(screen.getByLabelText('skill')).toBeInTheDocument();
  });

  it('removes an entry when remove button is clicked', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);

    expect(screen.getByText('#1')).toBeInTheDocument();

    const removeBtn = screen.getByRole('button', { name: /remove item 1/i });
    await userEvent.click(removeBtn);

    expect(screen.queryByText('#1')).not.toBeInTheDocument();
  });

  it('submits correct nested data with array indexing', async () => {
    const onSubmit = vi.fn();
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={onSubmit} />);

    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);

    const categoryInput = screen.getByLabelText('category');
    const skillInput = screen.getByLabelText('skill');
    await userEvent.type(categoryInput, 'Frontend');
    await userEvent.type(skillInput, 'React');

    const saveBtn = screen.getByRole('button', { name: /save/i });
    await userEvent.click(saveBtn);

    const data = onSubmit.mock.calls[0][0] as Record<string, unknown>;
    expect(data).toMatchObject({
      skills: [{ category: 'Frontend', skill: 'React' }],
    });
  });

  it('renders non-repeatable component as fieldset', () => {
    const schema: FieldDefinition[] = [
      {
        name: 'banner',
        type: 'component',
        fields: [{ name: 'title', type: 'text' }],
      },
    ];
    renderWithProviders(<ContentTypeBuilder schema={schema} mutationFn={noop} />);
    const group = screen.getByRole('group', { name: 'banner' });
    expect(group).toBeInTheDocument();
    expect(within(group).getByLabelText('title')).toBeInTheDocument();
  });
});
