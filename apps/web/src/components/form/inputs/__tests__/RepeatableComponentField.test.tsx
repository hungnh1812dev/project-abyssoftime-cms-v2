import { describe, it, expect, vi } from 'vitest';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test-utils';
import { ContentTypeBuilder } from '@/pages/admin/panels/content-type/ContentTypeBuilder';
import type { FieldDefinition } from '@/types/cms';

const noop = () => Promise.resolve();

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

async function addAndExpandEntry() {
  const addBtn = screen.getByRole('button', { name: /add entry/i });
  await userEvent.click(addBtn);
  const toggle = screen.getByRole('button', { expanded: false });
  await userEvent.click(toggle);
}

describe('RepeatableComponentField', () => {
  it('renders an add button when empty', () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    expect(screen.getByRole('button', { name: /add entry/i })).toBeInTheDocument();
  });

  it('adds an entry when add button is clicked', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    await addAndExpandEntry();

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

    await addAndExpandEntry();

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

describe('RepeatableComponentField — collapsible entries', () => {
  it('newly added entry is collapsed by default', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);

    expect(screen.getByText('#1')).toBeInTheDocument();
    expect(screen.queryByLabelText('category')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('skill')).not.toBeInTheDocument();
  });

  it('clicking entry header expands it', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    await addAndExpandEntry();

    expect(screen.getByLabelText('category')).toBeInTheDocument();
    expect(screen.getByLabelText('skill')).toBeInTheDocument();
  });

  it('clicking expanded entry header collapses it', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    await addAndExpandEntry();

    expect(screen.getByLabelText('category')).toBeInTheDocument();

    const toggle = screen.getByRole('button', { expanded: true });
    await userEvent.click(toggle);

    expect(screen.queryByLabelText('category')).not.toBeInTheDocument();
  });

  it('move/delete buttons visible when entry is collapsed', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);

    expect(screen.getByRole('button', { name: /move item 1 up/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /move item 1 down/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /remove item 1/i })).toBeInTheDocument();
  });

  it('hint text shows first text field value', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    await addAndExpandEntry();

    const categoryInput = screen.getByLabelText('category');
    await userEvent.type(categoryInput, 'Frontend');

    const toggle = screen.getByRole('button', { expanded: true });
    await userEvent.click(toggle);

    expect(screen.getByText(/Frontend/)).toBeInTheDocument();
  });

  it('expanding one entry does not affect others', async () => {
    renderWithProviders(<ContentTypeBuilder schema={repeatableSchema} mutationFn={noop} />);
    const addBtn = screen.getByRole('button', { name: /add entry/i });
    await userEvent.click(addBtn);
    await userEvent.click(addBtn);

    expect(screen.getByText('#1')).toBeInTheDocument();
    expect(screen.getByText('#2')).toBeInTheDocument();

    const toggles = screen.getAllByRole('button', { expanded: false });
    await userEvent.click(toggles[0]);

    expect(screen.getByLabelText('category')).toBeInTheDocument();

    const collapsedToggles = screen.getAllByRole('button', { expanded: false });
    expect(collapsedToggles.length).toBeGreaterThanOrEqual(1);
  });
});
