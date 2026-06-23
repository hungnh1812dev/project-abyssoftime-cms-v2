import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { AxiosError } from 'axios';
import { toast } from 'sonner';
import { api } from '@/lib/api';
import type { Locale } from '@/types/cms';

export function useCreateLocale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { code: string; name: string; isDefault: boolean }) => api.post<Locale>('/api/locales', data).then((response) => response.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['locales'] });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to create locale';
      toast.error(message);
    },
  });
}

export function useUpdateLocale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ code, ...data }: { code: string; name?: string; isDefault?: boolean }) => api.put<Locale>(`/api/locales/${code}`, data).then((response) => response.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['locales'] });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to update locale';
      toast.error(message);
    },
  });
}

export function useDeleteLocale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => api.delete(`/api/locales/${code}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['locales'] });
    },
    onError: (error: unknown) => {
      const message = (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to delete locale';
      toast.error(message);
    },
  });
}
