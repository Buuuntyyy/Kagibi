// Configuration
const IV_LENGTH = 12;
const TAG_LENGTH = 128; // bits

self.onmessage = async (e) => {
  const { type, fileChunk, key, chunkIndex, totalChunks } = e.data;

  try {
    if (type === 'ENCRYPT') {
      const result = await encryptChunk(fileChunk, key);
      // On renvoie le résultat au thread principal
      // On utilise 'transferable objects' ([result]) pour éviter la copie mémoire (très rapide)
      self.postMessage({ 
        type: 'ENCRYPT_SUCCESS', 
        encryptedChunk: result, 
        chunkIndex 
      }, [result]);

    } else if (type === 'DECRYPT') {
      const result = await decryptChunk(fileChunk, key);
      self.postMessage({ 
        type: 'DECRYPT_SUCCESS', 
        decryptedChunk: result, 
        chunkIndex 
      }, [result]);
    }
  } catch (error) {
    self.postMessage({ type: 'ERROR', error: error.message });
  }
};

async function encryptChunk(chunk, key) {
  const iv = crypto.getRandomValues(new Uint8Array(IV_LENGTH));
  const encryptedContent = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: iv, tagLength: TAG_LENGTH },
    key,
    chunk
  );

  // Concaténation IV + Data
  const combined = new Uint8Array(iv.byteLength + encryptedContent.byteLength);
  combined.set(iv);
  combined.set(new Uint8Array(encryptedContent), iv.byteLength);

  return combined.buffer; // Renvoie un ArrayBuffer
}

async function decryptChunk(chunk, key) {
  // chunk est un ArrayBuffer
  const iv = chunk.slice(0, IV_LENGTH);
  const data = chunk.slice(IV_LENGTH);

  const decryptedContent = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: new Uint8Array(iv), tagLength: TAG_LENGTH },
    key,
    data
  );

  return decryptedContent;
}