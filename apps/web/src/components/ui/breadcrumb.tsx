import { Link } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';
import type { BreadcrumbItem } from '@/hooks/useBreadcrumbs';

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

export function Breadcrumb({ items }: BreadcrumbProps) {
  return (
    <nav aria-label="Breadcrumb">
      <ol className="flex items-center gap-1.5 text-sm">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;
          return (
            <li key={index} className="flex items-center gap-1.5">
              {index > 0 && <ChevronRight className="text-muted-foreground size-3.5" />}
              {item.to && !isLast ? (
                <Link to={item.to} className="text-muted-foreground hover:text-foreground transition-colors">
                  {item.label}
                </Link>
              ) : (
                <span className={isLast ? 'text-foreground font-medium' : 'text-muted-foreground'}>{item.label}</span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
