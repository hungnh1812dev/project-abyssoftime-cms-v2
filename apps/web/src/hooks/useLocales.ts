import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'

export function useLocales() {
  return useQuery({
    queryKey: ['locales'] as const,
    queryFn: () => api.get<string[]>('/api/locales').then((response) => response.data),
  })
}
