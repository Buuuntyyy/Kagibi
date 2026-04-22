/**
 * Reusable pool of crypto workers.
 * Workers are created once at module load and reused across all encrypt/decrypt calls,
 * eliminating the per-chunk thread spawn/teardown overhead that causes high CPU in the profiler.
 */

const POOL_SIZE = Math.min(Math.max((navigator.hardwareConcurrency ?? 4), 2), 8)

class CryptoWorkerPool {
  constructor() {
    this._idle = []
    this._queue = []

    for (let i = 0; i < POOL_SIZE; i++) {
      const w = new Worker(new URL('./crypto.worker.js', import.meta.url), { type: 'module' })
      w._task = null
      w.onmessage = (e) => this._onMessage(w, e)
      w.onerror = (e) => this._onError(w, e)
      this._idle.push(w)
    }
  }

  _onMessage(worker, e) {
    const task = worker._task
    worker._task = null
    this._idle.push(worker)

    if (task) {
      const { type, encryptedChunk, decryptedChunk, error } = e.data
      if (type === 'ERROR') task.reject(new Error(error))
      else if (type === 'ENCRYPT_SUCCESS') task.resolve(encryptedChunk)
      else if (type === 'DECRYPT_SUCCESS') task.resolve(decryptedChunk)
    }

    this._dispatch()
  }

  _onError(worker, e) {
    const task = worker._task
    worker._task = null
    this._idle.push(worker)
    if (task) task.reject(new Error(e.message || 'Worker error'))
    this._dispatch()
  }

  _dispatch() {
    if (this._queue.length === 0 || this._idle.length === 0) return
    const worker = this._idle.pop()
    const task = this._queue.shift()
    worker._task = task
    worker.postMessage(task.msg, task.transfer)
  }

  /**
   * Send a message to an available worker, queuing if all are busy.
   * @param {object} msg  - Message payload for crypto.worker.js
   * @param {Transferable[]} transfer - Transferable objects (e.g. ArrayBuffers)
   * @returns {Promise<ArrayBuffer>}
   */
  run(msg, transfer = []) {
    return new Promise((resolve, reject) => {
      this._queue.push({ msg, transfer, resolve, reject })
      this._dispatch()
    })
  }
}

export const cryptoWorkerPool = new CryptoWorkerPool()
