import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'

export interface InviteItem {
  id: string
  email: string
  role: string
  expiresAt: string
  createdBy: string
  createdAt: string
}

interface InviteListResponse {
  items: InviteItem[]
}

interface CreateInviteResponse {
  id: string
  email: string
  role: string
  expiresAt: string
  token: string
}

const KEYS = {
  list: ['invites'] as const,
}

export function useInviteList() {
  return useQuery<InviteListResponse>({
    queryKey: KEYS.list,
    queryFn: () => api.get<InviteListResponse>('/api/invites').then((r) => r.data),
  })
}

export function useCreateInvite() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) =>
      api.post<CreateInviteResponse>('/api/invites', { email, role }).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.list })
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to create invite'
      toast.error(msg)
    },
  })
}

export function useRevokeInvite() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/invites/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.list })
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to revoke invite'
      toast.error(msg)
    },
  })
}

export function useAcceptInvite() {
  return useMutation({
    mutationFn: ({ token, password }: { token: string; password: string }) =>
      api.post(`/auth/invite/${token}`, { password }).then((r) => r.data),
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to accept invite'
      toast.error(msg)
    },
  })
}
