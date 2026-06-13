import { createContext, useContext } from 'react'

interface FormStateContextValue {
  loading: boolean
  submitting: boolean
}

export const FormStateContext = createContext<FormStateContextValue>({
  loading: false,
  submitting: false,
})

export function useCmsFormState() {
  return useContext(FormStateContext)
}
