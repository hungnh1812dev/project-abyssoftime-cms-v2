import { useForm, FormProvider as RHFFormProvider } from 'react-hook-form'
import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { FormStateContext } from './FormStateContext'

interface CmsFormProviderProps {
  query?: UseQueryOptions
  mutationFn: (data: Record<string, unknown>) => Promise<unknown>
  onSuccess?: () => void
  children: React.ReactNode
}

export function FormProvider({ query, mutationFn, onSuccess, children }: CmsFormProviderProps) {
  const queryClient = useQueryClient()

  const { data, isFetching } = useQuery(
    query ?? { queryKey: ['__noop__'], queryFn: () => null, enabled: false },
  )

  const methods = useForm({
    values: (data as Record<string, unknown>) ?? {},
  })

  const { isDirty } = methods.formState

  const { mutate, isPending } = useMutation({
    mutationFn,
    onSuccess: () => {
      toast.success('Saved')
      methods.reset(methods.getValues())
      if (query?.queryKey) {
        queryClient.invalidateQueries({ queryKey: query.queryKey as readonly unknown[] })
      }
      onSuccess?.()
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Something went wrong'
      toast.error(msg)
    },
  })

  function onSubmit(values: Record<string, unknown>) {
    mutate(values)
  }

  return (
    <FormStateContext.Provider value={{ loading: isFetching, submitting: isPending, isDirty }}>
      <RHFFormProvider {...methods}>
        <form onSubmit={methods.handleSubmit(onSubmit)}>{children}</form>
      </RHFFormProvider>
    </FormStateContext.Provider>
  )
}
