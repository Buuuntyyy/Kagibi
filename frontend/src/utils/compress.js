/**
 * Client-side compression utilities using the native CompressionStream / DecompressionStream API.
 * Compression is applied BEFORE encryption so ciphertext (which is random and incompressible)
 * is never sent through a compressor.
 *
 * Only compressible MIME types are targeted. Binary formats that are already compressed
 * (JPEG, MP4, ZIP, PNG…) are skipped since they would not shrink.
 */

const compressiblePrefixes = [
  'text/',
  'application/json',
  'application/xml',
  'application/javascript',
  'application/x-javascript',
  'application/svg+xml',
  'application/csv',
  'application/vnd.openxmlformats-officedocument', // .docx .xlsx .pptx
  'application/vnd.oasis.opendocument',             // .odt .ods .odp
  'application/rtf',
]

const nonCompressibleExts = new Set([
  'jpg', 'jpeg', 'png', 'gif', 'webp', 'avif', 'heic',
  'mp4', 'mp3', 'aac', 'ogg', 'flac', 'opus', 'wav',
  'mov', 'avi', 'mkv', 'wmv', 'webm',
  'zip', 'gz', 'bz2', 'xz', '7z', 'rar', 'zst',
  'pdf',
])

/**
 * Whether the file with the given MIME type and name is worth compressing.
 * @param {string} mimeType
 * @param {string} filename
 * @returns {boolean}
 */
export function shouldCompress(mimeType, filename) {
  const ext = (filename.split('.').pop() || '').toLowerCase()
  if (nonCompressibleExts.has(ext)) return false
  const mt = (mimeType || '').toLowerCase()
  return compressiblePrefixes.some(p => mt.startsWith(p))
}

/**
 * Compress a Blob/File using gzip via CompressionStream.
 * Falls back to the original blob if CompressionStream is unavailable.
 * @param {Blob|File} blob
 * @returns {Promise<Blob>} compressed blob (content-type: application/octet-stream)
 */
export async function compressBlob(blob) {
  if (typeof CompressionStream === 'undefined') {
    return blob
  }
  const cs = new CompressionStream('gzip')
  const compressed = blob.stream().pipeThrough(cs)
  return new Response(compressed).blob()
}

/**
 * Decompress a gzip Blob using DecompressionStream.
 * @param {Blob} blob  - gzip-compressed data
 * @param {string} mimeType - target MIME type for the resulting Blob
 * @returns {Promise<Blob>}
 */
export async function decompressBlob(blob, mimeType = 'application/octet-stream') {
  if (typeof DecompressionStream === 'undefined') {
    throw new Error('DecompressionStream not supported in this browser')
  }
  const ds = new DecompressionStream('gzip')
  const decompressed = blob.stream().pipeThrough(ds)
  const result = await new Response(decompressed).blob()
  return new Blob([result], { type: mimeType })
}
