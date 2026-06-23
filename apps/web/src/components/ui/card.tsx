import { cn } from '@/lib/utils';

function Card({ className, children, ...props }: React.ComponentProps<'div'>) {
  return (
    <div className={cn('bg-card border-border rounded-lg border shadow-sm', className)} {...props}>
      {children}
    </div>
  );
}

function CardHeader({ className, children, ...props }: React.ComponentProps<'div'>) {
  return (
    <div className={cn('border-border border-b px-6 py-4', className)} {...props}>
      {children}
    </div>
  );
}

function CardContent({ className, children, ...props }: React.ComponentProps<'div'>) {
  return (
    <div className={cn('p-6', className)} {...props}>
      {children}
    </div>
  );
}

export { Card, CardHeader, CardContent };
