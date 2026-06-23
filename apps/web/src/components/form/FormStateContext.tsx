import { createContext, useContext } from 'react';

interface FormStateContextValue {
  loading: boolean;
  submitting: boolean;
  isDirty: boolean;
}

export const FormStateContext = createContext<FormStateContextValue>({
  loading: false,
  submitting: false,
  isDirty: false,
});

export function useCmsFormState() {
  return useContext(FormStateContext);
}
