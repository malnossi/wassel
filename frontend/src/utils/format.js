/**
 * Format bytes into a human-readable size string.
 * @param {number} bytes - The number of bytes.
 * @param {number} [decimals=2] - Number of decimal places.
 * @returns {string}
 */
export const formatBytes = (bytes, decimals = 2) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

/**
 * Format transfer speed in bytes/sec to a human-readable string.
 * @param {number} bytesPerSec - Bytes per second.
 * @returns {string}
 */
export const formatSpeed = (bytesPerSec) => {
  if (!bytesPerSec || bytesPerSec <= 0) return '0 B/s'
  return formatBytes(bytesPerSec, 1) + '/s'
}
