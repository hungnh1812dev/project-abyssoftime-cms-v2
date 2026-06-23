import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test-utils';
import { ColumnChooserDialog } from '../ColumnChooserDialog';
import type { ContentType } from '@/types/cms';

const contentType: ContentType = {
  ID: 'ct-1',
  Name: 'Blog Posts',
  Slug: 'blog-posts',
  Kind: 'collection',
  Fields: [
    { name: 'title', type: 'text' },
    { name: 'slug', type: 'text' },
    { name: 'body', type: 'richtext' },
    { name: 'featured', type: 'boolean' },
    { name: 'banner', type: 'component' },
  ],
  CreatedAt: '',
  UpdatedAt: '',
};

function renderDialog(props: Partial<React.ComponentProps<typeof ColumnChooserDialog>> = {}) {
  const defaults = {
    open: true,
    onOpenChange: vi.fn(),
    contentType,
    currentListFields: [] as string[],
    onSave: vi.fn(),
    isSaving: false,
    ...props,
  };
  return { ...renderWithProviders(<ColumnChooserDialog {...defaults} />), ...defaults };
}

describe('ColumnChooserDialog', () => {
  it('renders content fields excluding component types', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText('title')).toBeInTheDocument();
      expect(screen.getByText('slug')).toBeInTheDocument();
      expect(screen.getByText('body')).toBeInTheDocument();
      expect(screen.getByText('featured')).toBeInTheDocument();
      expect(screen.queryByText('banner')).not.toBeInTheDocument();
    });
  });

  it('renders system fields section', async () => {
    renderDialog();
    await waitFor(() => {
      expect(screen.getByText('Created At')).toBeInTheDocument();
      expect(screen.getByText('Updated At')).toBeInTheDocument();
      expect(screen.getByText('Updated By')).toBeInTheDocument();
    });
  });

  it('defaults to first 3 content fields + all system fields when currentListFields is empty', async () => {
    renderDialog({ currentListFields: [] });
    await waitFor(() => {
      const checkboxes = screen.getAllByRole('checkbox');
      const checkedNames: string[] = [];
      for (const checkbox of checkboxes) {
        if ((checkbox as HTMLInputElement).checked) {
          const label = checkbox.closest('label');
          if (label) checkedNames.push(label.textContent ?? '');
        }
      }
      expect(checkedNames).toContain('title');
      expect(checkedNames).toContain('slug');
      expect(checkedNames).toContain('body');
      expect(checkedNames).not.toContain('featured');
      expect(checkedNames).toContain('Created At');
      expect(checkedNames).toContain('Updated At');
      expect(checkedNames).toContain('Updated By');
    });
  });

  it('initializes from currentListFields when provided', async () => {
    renderDialog({ currentListFields: ['title', 'createdAt'] });
    await waitFor(() => {
      const checkboxes = screen.getAllByRole('checkbox');
      const checkedNames: string[] = [];
      for (const checkbox of checkboxes) {
        if ((checkbox as HTMLInputElement).checked) {
          const label = checkbox.closest('label');
          if (label) checkedNames.push(label.textContent ?? '');
        }
      }
      expect(checkedNames).toEqual(['title', 'Created At']);
    });
  });

  it('calls onSave with selected fields in correct order', async () => {
    const user = userEvent.setup();
    const { onSave } = renderDialog({ currentListFields: ['title', 'slug', 'createdAt'] });

    await waitFor(() => screen.getByText('featured'));
    await user.click(screen.getByText('featured'));
    await user.click(screen.getByText('Save'));

    expect(onSave).toHaveBeenCalledWith(['title', 'slug', 'featured', 'createdAt']);
  });

  it('calls onOpenChange(false) when Cancel is clicked', async () => {
    const user = userEvent.setup();
    const { onOpenChange } = renderDialog();

    await waitFor(() => screen.getByText('Cancel'));
    await user.click(screen.getByText('Cancel'));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
