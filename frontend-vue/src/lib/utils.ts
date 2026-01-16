import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}


export function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
}

export function formatCompactNumber(number: number): string {
  if (number >= 1000000000) {
    return (number / 1000000000).toLocaleString('id-ID', { maximumFractionDigits: 2 }) + ' M';
  }
  if (number >= 1000000) {
    return (number / 1000000).toLocaleString('id-ID', { maximumFractionDigits: 2 }) + ' Jt';
  }
  if (number >= 1000) {
    return (number / 1000).toLocaleString('id-ID', { maximumFractionDigits: 2 }) + ' Rb';
  }
  return number.toString();
}
