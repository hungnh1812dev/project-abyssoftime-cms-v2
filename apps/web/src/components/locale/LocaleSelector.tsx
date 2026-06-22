import { useLocales } from '@/hooks/useLocales';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface LocaleSelectorProps {
  value: string;
  onChange: (code: string) => void;
}

export function LocaleSelector({ value, onChange }: LocaleSelectorProps) {
  const { data: locales = [] } = useLocales();

  const resolvedValue = value || locales.find((loc) => loc.isDefault)?.code || locales[0]?.code || '';

  if (locales.length === 0) return null;

  return (
    <Select value={resolvedValue} onValueChange={(code) => onChange(code || '')}>
      <SelectTrigger size="sm" className="w-40">
        <SelectValue placeholder="Select locale">
          {locales.find((loc) => loc.code === resolvedValue)?.name ?? resolvedValue}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {locales.map((locale) => (
          <SelectItem key={locale.code} value={locale.code}>
            {locale.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
