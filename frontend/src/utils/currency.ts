export function formatUSD(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

export function formatKHR(amount: number): string {
  return new Intl.NumberFormat('km-KH', {
    style: 'currency',
    currency: 'KHR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function convertToKHR(usdAmount: number, exchangeRate: number): number {
  return usdAmount * exchangeRate
}
