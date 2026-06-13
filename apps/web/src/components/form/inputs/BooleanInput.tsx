import { Controller, type Control } from 'react-hook-form'
import { Switch } from '@/components/ui/switch'

interface BooleanInputProps {
  name?: string
  control?: Control
  'aria-label'?: string
}

export function BooleanInput({ name, control, 'aria-label': ariaLabel }: BooleanInputProps) {
  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue={false}
      render={({ field }) => (
        <Switch
          checked={field.value as boolean}
          onCheckedChange={field.onChange}
          aria-label={ariaLabel}
        />
      )}
    />
  )
}
