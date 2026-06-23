import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { AxiosError } from 'axios';
import { toast } from 'sonner';
import { api } from '@/lib/api';

export interface InviteItem {
  id: string;
  email: string;
  role: string;
  expiresAt: string;
  createdBy: string;
  createdAt: string;
}

interface InviteListResponse {
  items: InviteItem[];
}

interface CreateInviteResponse {
  id: string;
  email: string;
  role: string;
  expiresAt: string;
  token: string;
}

const KEYS = {
  list: ['invites'] as const,
};

export function useInviteList() {
  return useQuery<InviteListResponse>({
    queryKey: KEYS.list,
    queryFn: () => api.get<InviteListResponse>('/api/invites').then((response) => response.data),
  });
}

export function useCreateInvite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) => api.post<CreateInviteResponse>('/api/invites', { email, role }).then((response) => response.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: KEYS.list });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to create invite';
      toast.error(message);
    },
  });
}

export function useRevokeInvite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/invites/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: KEYS.list });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to revoke invite';
      toast.error(message);
    },
  });
}

export function useAcceptInvite() {
  return useMutation({
    mutationFn: ({ token, password, displayName }: { token: string; password: string; displayName: string }) =>
      api.post(`/auth/invite/${token}`, { password, displayName }).then((response) => response.data),
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to accept invite';
      toast.error(message);
    },
  });
}
