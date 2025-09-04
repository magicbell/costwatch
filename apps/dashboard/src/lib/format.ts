const number = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 0,
  maximumFractionDigits: 2,
});

const currency = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
  currencyDisplay: 'symbol',
});

const time = new Intl.DateTimeFormat('en-US', {
  hour: '2-digit',
  minute: '2-digit',
  hour12: true,
});

const date = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
});

const datetime = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: true,
});

export function formatNumber(
  value: number | null | undefined,
  opts?: { minimumFractionDigits?: number; maximumFractionDigits?: number },
): string {
  const v = typeof value === 'number' && isFinite(value) ? value : 0;
  if (opts) {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: opts.minimumFractionDigits ?? 0,
      maximumFractionDigits: opts.maximumFractionDigits ?? 2,
    }).format(v);
  }
  return number.format(v);
}

export function formatCurrency(value: number | null | undefined): string {
  const v = typeof value === 'number' && isFinite(value) ? value : 0;
  return currency.format(v);
}

export function formatTime(value: string | number | Date): string {
  const v = new Date(value);
  return time.format(v);
}

export function formatDateTime(value: string | number | Date): string {
  const v = new Date(value);
  return datetime.format(v);
}

export function formatDate(value: string | number | Date): string {
  const v = new Date(value);
  return date.format(v);
}
