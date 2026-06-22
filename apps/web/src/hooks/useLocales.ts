import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Locale } from '@/types/cms'

export function useLocales() {
  return useQuery({
    queryKey: ['locales'] as const,
    queryFn: () => api.get<Locale[]>('/api/locales').then((response) => response.data),
  })
}
