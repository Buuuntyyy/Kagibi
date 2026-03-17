export const formatSize = (bytes) => {
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  if (bytes === 0) return '0 Byte'
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export const formatSpeed = (bytesPerSecond) => {
  const value = Number(bytesPerSecond) || 0
  if (value <= 0) return '0 B/s'

  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s', 'TB/s']
  const k = 1024
  const i = Math.min(Math.floor(Math.log(value) / Math.log(k)), units.length - 1)
  return Number.parseFloat((value / Math.pow(k, i)).toFixed(2)) + ' ' + units[i]
}
