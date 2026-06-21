import { useState } from 'react';
import { Controller, type Control } from 'react-hook-form';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';

interface JsonInputProps {
  name?: string;
  control?: Control;
}

function serialize(value: unknown): string {
  if (value == null) return '';
  return JSON.stringify(value, null, 2);
}

function InnerJsonInput({ field }: { field: { value: unknown; onChange: (v: unknown) => void } }) {
  const [rawValue, setRawValue] = useState(serialize(field.value));
  const [syntaxError, setSyntaxError] = useState<string | null>(null);
  const [editCount, setEditCount] = useState(0);
  const [syncedAt, setSyncedAt] = useState(0);

  const [prevSerialized, setPrevSerialized] = useState(() => serialize(field.value));
  const currentSerialized = serialize(field.value);

  if (currentSerialized !== prevSerialized) {
    setPrevSerialized(currentSerialized);
    if (editCount === syncedAt) {
      setRawValue(currentSerialized);
    }
    setSyncedAt(editCount);
  }

  return (
    <div>
      <div data-testid="json-editor-wrapper" className="border-input min-h-[15em] overflow-hidden rounded-md border">
        <CodeMirror
          value={rawValue}
          extensions={[json()]}
          minHeight="15em"
          onChange={(val) => {
            setRawValue(val);
            setEditCount((c) => c + 1);
            if (val.trim() === '') {
              setSyntaxError(null);
              field.onChange(null);
              return;
            }
            try {
              const parsed = JSON.parse(val);
              setSyntaxError(null);
              field.onChange(parsed);
            } catch {
              setSyntaxError('Invalid JSON');
              field.onChange(undefined);
            }
          }}
        />
      </div>
      {syntaxError && <p role="alert">{syntaxError}</p>}
    </div>
  );
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
  );
}
