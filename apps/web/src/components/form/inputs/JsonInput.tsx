import { useState } from 'react'
import { Controller, type Control } from 'react-hook-form'
import CodeMirror from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'

interface JsonInputProps {
  name?: string
  control?: Control
}

export function JsonInput({ name, control }: JsonInputProps) {
  const [rawValue, setRawValue] = useState('')
  const [syntaxError, setSyntaxError] = useState<string | null>(null)

  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue={null}
      rules={{
        validate: (v) => v !== undefined || 'Invalid JSON',
      }}
      render={({ field }) => (
        <div>
          <div data-testid="json-editor-wrapper" className="min-h-[15em]">
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
      )}
    />
  )
}
