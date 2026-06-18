import { type ComponentProps } from 'react'
import { Input } from '@/components/ui/input'

interface NumberInputProps extends Omit<ComponentProps<'input'>, 'type'> {
  step?: number
  min?: number
  max?: number
}

export function NumberInput({ step, min, max, ...props }: NumberInputProps) {
  return <Input type="number" step={step} min={min} max={max} {...props} />
}
