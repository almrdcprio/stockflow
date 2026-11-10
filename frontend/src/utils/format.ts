export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(value)
}

export function formatRelativeTime(
  dateString: string | null,
): string {
  if (!dateString) return 'No History'

  const date = new Date(dateString)
  const diff = Date.now() - date.getTime()

  if (Number.isNaN(date.getTime())) return 'Unknown'

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 30) return `${days}d ago`

  return date.toLocaleDateString('id-ID')
}