import { ButtonHTMLAttributes } from 'react';
import clsx from 'clsx';

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'outline';
};

export function Button({ className, variant = 'primary', ...props }: ButtonProps) {
  const variantClass = {
    primary: 'bg-primary text-white hover:bg-primary/90',
    secondary: 'border border-slate-300 text-slate-700 hover:bg-slate-100',
    ghost: 'text-slate-600 hover:bg-slate-100',
    outline: 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50'
  }[variant];

  return (
    <button
      className={clsx(
        'rounded-md px-4 py-2 text-sm font-semibold transition-colors disabled:opacity-70',
        variantClass,
        className
      )}
      {...props}
    />
  );
}
