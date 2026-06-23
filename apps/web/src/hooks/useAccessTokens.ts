import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { AxiosError } from 'axios';
import { toast } from 'sonner';
import { api } from '@/lib/api';

export interface AccessTokenItem {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdAt: string;
}

interface TokenListResponse {
  items: AccessTokenItem[];
  total: number;
  page: number;
  limit: number;
}

interface CreateTokenResponse {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  expiresAt: string | null;
  createdAt: string;
  token: string;
}

const KEYS = {
  list: (page: number) => ['access-tokens', 'list', page] as const,
  all: ['access-tokens'] as const,
};

export function useAccessTokenList(page: number) {
  return useQuery<TokenListResponse>({
    queryKey: KEYS.list(page),
    queryFn: () => api.get<TokenListResponse>(`/api/access-tokens?page=${page}&limit=20`).then((response) => response.data),
  });
}

export function useCreateAccessToken() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; scopes: string[]; expiresIn?: string }) => api.post<CreateTokenResponse>('/api/access-tokens', data).then((response) => response.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: KEYS.all });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to create token';
      toast.error(message);
    },
  });
}

export function useDeleteAccessToken() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/access-tokens/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: KEYS.all });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to delete token';
      toast.error(message);
    },
  });
}
