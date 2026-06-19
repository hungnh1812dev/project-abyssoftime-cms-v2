import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'

export interface UserItem {
  id: string
  email: string
  role: string
}

interface UserListResponse {
  items: UserItem[]
  total: number
  page: number
  limit: number
}

const KEYS = {
  list: (page: number) => ['users', 'list', page] as const,
  all: ['users'] as const,
}

export function useUserList(page: number) {
  return useQuery<UserListResponse>({
    queryKey: KEYS.list(page),
    queryFn: () =>
      api.get<UserListResponse>(`/api/users?page=${page}&limit=20`).then((r) => r.data),
  })
}

export function useUpdateUserRole() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) =>
      api.put(`/api/users/${id}/role`, { role }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to update role'
      toast.error(msg)
    },
  })
}

export function useDeleteUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/users/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Failed to delete user'
      toast.error(msg)
    },
  })
}
