import { useForm, FormProvider as RHFFormProvider } from 'react-hook-form'
import { useQuery, useMutation, type UseQueryOptions } from '@tanstack/react-query'
import { FormStateContext } from './FormStateContext'

interface CmsFormProviderProps {
  query?: UseQueryOptions
  mutationFn: (data: Record<string, unknown>) => Promise<unknown>
  onSuccess?: () => void
  children: React.ReactNode
}

export function FormProvider({ query, mutationFn, onSuccess, children }: CmsFormProviderProps) {
  const { data, isFetching } = useQuery(
    query ?? { queryKey: ['__noop__'], queryFn: () => null, enabled: false },
  )

  const { mutate, isPending } = useMutation({ mutationFn })

  const methods = useForm({
    values: (data as Record<string, unknown>) ?? {},
  })

  function onSubmit(values: Record<string, unknown>) {
    mutate(values, { onSuccess })
  }

  return (
    <FormStateContext.Provider value={{ loading: isFetching, submitting: isPending }}>
      <RHFFormProvider {...methods}>
        <form onSubmit={methods.handleSubmit(onSubmit)}>{children}</form>
      </RHFFormProvider>
    </FormStateContext.Provider>
  )
}
