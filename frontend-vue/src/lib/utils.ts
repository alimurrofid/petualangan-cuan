import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import { format } from "date-fns"
import { id as idLocale } from "date-fns/locale"

export function formatDate(dateStr: string | null | undefined, fmt = 'dd MM yyyy'): string {
  if (!dateStr) return '-';
  try {
    return format(new Date(dateStr), fmt, { locale: idLocale });
  } catch {
    return '-';
  }
}

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}


import { useSettingsStore } from "@/stores/settings";

export function formatCurrency(value: number | string | undefined | null, options?: { showDecimal?: boolean }) {
  let showDecimal = options?.showDecimal;
  if (showDecimal === undefined) {
    try {
      const settingsStore = useSettingsStore();
      showDecimal = settingsStore.showDecimal;
    } catch {
      try {
        const raw = localStorage.getItem('cuan_settings');
        if (raw) {
          const parsed = JSON.parse(raw);
          showDecimal = parsed.showDecimal === true || parsed.showDecimal === 'true' || parsed.showDecimal === 'Show';
        }
      } catch {
        showDecimal = false;
      }
    }
  }

  const num = typeof value === 'number' && !isNaN(value) ? value : Number(value) || 0;
  const fractionDigits = showDecimal ? 2 : 0;
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(num);
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

export function parseCurrencyInput(value: string): number {
  if (!value) return 0;
  let clean = value.replace(/[^0-9,]/g, "");
  const normalized = clean.replace(",", ".");
  return parseFloat(normalized) || 0;
}

export function formatCurrencyInput(value: number | string): string {
  if (!value) return "";
  const num = Number(value);
  if (isNaN(num)) return "";
  const formatted = new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(num);

  return `Rp ${formatted}`;
}

export function formatCurrencyLive(value: string): string {
  let clean = value.replace(/[^0-9,]/g, "");
  const parts = clean.split(',');
  if (parts.length > 0 && typeof parts[0] === 'string') {
      parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ".");
  }
  let formatted = parts[0] || "";
  if (parts.length > 1) {
      const decimal = parts[1];
      if (typeof decimal === 'string') {
          formatted += `,${decimal.slice(0, 2)}`;
      }
  } else if (value.includes(',') || (parts.length === 1 && clean.endsWith(','))) {
     formatted += `,`;
  }
  return formatted ? `Rp ${formatted}` : "";
}
