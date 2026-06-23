import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { FormProvider } from '../../FormProvider';
import { FormField } from '../../FormField';
import { MediaInput } from '../MediaInput';
import { api } from '@/lib/api';

let mock: MockAdapter;

const mediaResponse = {
  items: [
    {
      ID: 'a1',
      documentId: 'doc-uuid-1',
      url: 'https://cdn/a1.jpg',
      thumbnailUrl: 'https://cdn/a1.jpg',
      publicId: 'p1',
      fileName: 'a1_abc.jpg',
      fileExt: 'jpg',
      hash: 'abc',
      width: 800,
      height: 600,
      createdAt: '',
    },
  ],
  total: 1,
  page: 1,
  limit: 20,
};

beforeEach(() => {
  mock = new MockAdapter(api);
  mock.onGet('/api/media?page=1&limit=20').reply(200, mediaResponse);
});

afterEach(() => {
  mock.restore();
  vi.clearAllMocks();
});

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={createClient()}>{children}</QueryClientProvider>;
}

describe('MediaInput', () => {
  it('renders a clickable upload zone with placeholder text', () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    );
    expect(screen.getByTestId('media-upload-zone')).toBeInTheDocument();
    expect(screen.getByText(/click to select media/i)).toBeInTheDocument();
  });

  it('opens the MediaLibrary dialog when zone is clicked', async () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    );

    await userEvent.click(screen.getByTestId('media-upload-zone'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('stores documentId and shows preview URL when an asset is selected', async () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    );

    await userEvent.click(screen.getByTestId('media-upload-zone'));

    await waitFor(() => expect(screen.getAllByRole('img')).toHaveLength(1));
    await userEvent.click(screen.getByRole('img', { name: mediaResponse.items[0].fileName }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    const zoneImg = await screen.findByRole('img', { name: mediaResponse.items[0].fileName });
    expect(zoneImg).toHaveAttribute('src', 'https://cdn/a1.jpg');
  });

  it('closes the library without changing value when Close is clicked', async () => {
    render(
      <Wrapper>
        <FormProvider mutationFn={vi.fn().mockResolvedValue(undefined)}>
          <FormField name="image">
            <MediaInput />
          </FormField>
        </FormProvider>
      </Wrapper>,
    );

    await userEvent.click(screen.getByTestId('media-upload-zone'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: /close/i }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(screen.queryByRole('img', { name: /media preview/i })).not.toBeInTheDocument();
  });
});
