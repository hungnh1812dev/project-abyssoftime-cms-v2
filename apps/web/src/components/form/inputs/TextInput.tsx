import { type ComponentProps } from 'react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

interface TextInputProps extends Omit<ComponentProps<'input'>, 'type'> {
  multiline?: boolean
}

export function TextInput({ multiline, ...props }: TextInputProps) {
  if (multiline) {
    return <Textarea {...(props as ComponentProps<'textarea'>)} />
  }
  return <Input type="text" {...props} />
}
