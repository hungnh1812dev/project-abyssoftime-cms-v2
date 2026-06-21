import { Controller, type Control } from 'react-hook-form'
import { CKEditor } from '@ckeditor/ckeditor5-react'
import ClassicEditor from '@ckeditor/ckeditor5-build-classic'

interface RichTextInputProps {
  name?: string
  control?: Control
  toolbar?: string[]
}

// @ckeditor/ckeditor5-build-classic v41 types predate the react-wrapper v11
// expectation of v42+. The runtime API is compatible; cast to satisfy tsc.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const Editor = ClassicEditor as any

const minHeightStyle = '.ck-editor__editable_inline { min-height: 12em; }'

export function RichTextInput({ name, control, toolbar }: RichTextInputProps) {
  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue=""
      render={({ field }) => (
        <>
          <style>{minHeightStyle}</style>
          <CKEditor
            editor={Editor}
            data={(field.value as string) ?? ''}
            config={toolbar ? { toolbar } : undefined}
            onChange={(_event, editor) => {
              field.onChange(editor.getData())
            }}
          />
        </>
      )}
    />
  )
}
