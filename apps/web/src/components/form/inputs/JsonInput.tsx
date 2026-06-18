import { useState, useEffect } from 'react'
import { Controller, type Control } from 'react-hook-form'
import CodeMirror from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'

interface JsonInputProps {
  name?: string
  control?: Control
}

function serialize(value: unknown): string {
  if (value == null) return ''
  return JSON.stringify(value, null, 2)
}

function InnerJsonInput({ field }: { field: { value: unknown; onChange: (v: unknown) => void } }) {
  const [rawValue, setRawValue] = useState(() => serialize(field.value))
  const [syntaxError, setSyntaxError] = useState<string | null>(null)

  useEffect(() => {
    setRawValue(serialize(field.value))
  }, [field.value])

  return (
    <div>
      <div data-testid="json-editor-wrapper" className="min-h-[15em] border border-input rounded-md overflow-hidden">
        <CodeMirror
          value={rawValue}
          extensions={[json()]}
          minHeight="15em"
          onChange={(val) => {
            setRawValue(val)
            if (val.trim() === '') {
              setSyntaxError(null)
              field.onChange(null)
              return
            }
            try {
              const parsed = JSON.parse(val)
              setSyntaxError(null)
              field.onChange(parsed)
            } catch {
              setSyntaxError('Invalid JSON')
              field.onChange(undefined)
            }
          }}
        />
      </div>
      {syntaxError && <p role="alert">{syntaxError}</p>}
    </div>
  )
}

export function JsonInput({ name, control }: JsonInputProps) {
  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue={null}
      rules={{
        validate: (v) => v !== undefined || 'Invalid JSON',
      }}
      render={({ field }) => <InnerJsonInput field={field} />}
    />
  )
}
