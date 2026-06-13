import React from 'react'
import { useFormContext, type FieldError } from 'react-hook-form'

interface FormFieldProps {
  name: string
  children: React.ReactElement<Record<string, unknown>>
}

function getNestedError(errors: Record<string, unknown>, path: string): FieldError | undefined {
  return path.split('.').reduce((acc: unknown, key) => {
    if (acc && typeof acc === 'object') return (acc as Record<string, unknown>)[key]
    return undefined
  }, errors) as FieldError | undefined
}

export function FormField({ name, children }: FormFieldProps) {
  const { register, control, formState } = useFormContext()
  const error = getNestedError(formState.errors as Record<string, unknown>, name)

  const child = React.cloneElement(children, {
    ...register(name),
    control,
    name,
  })

  return (
    <div>
      {child}
      {error?.message && (
        <p role="alert" aria-label={`${name}-error`}>
          {String(error.message)}
        </p>
      )}
    </div>
  )
}
