import { Controller, type Control } from 'react-hook-form'
import { CKEditor } from '@ckeditor/ckeditor5-react'
import {
  ClassicEditor,
  Essentials,
  Paragraph,
  Bold,
  Italic,
  Heading,
  Link,
  List,
  BlockQuote,
  Indent,
  MediaEmbed,
  Table,
  TableToolbar,
} from 'ckeditor5'
import 'ckeditor5/ckeditor5.css'

interface RichTextInputProps {
  name?: string
  control?: Control
  toolbar?: string[]
}

const DEFAULT_TOOLBAR = [
  'heading', '|',
  'bold', 'italic', 'link', '|',
  'bulletedList', 'numberedList', '|',
  'outdent', 'indent', '|',
  'blockQuote', 'insertTable', 'mediaEmbed', '|',
  'undo', 'redo',
]

const PLUGINS = [
  Essentials, Paragraph, Bold, Italic, Heading, Link,
  List, BlockQuote, Indent, MediaEmbed, Table, TableToolbar,
]

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
            editor={ClassicEditor}
            data={(field.value as string) ?? ''}
            config={{
              licenseKey: 'GPL',
              plugins: PLUGINS,
              toolbar: toolbar ?? DEFAULT_TOOLBAR,
            }}
            onChange={(_event, editor) => {
              field.onChange(editor.getData())
            }}
          />
        </>
      )}
    />
  )
}
