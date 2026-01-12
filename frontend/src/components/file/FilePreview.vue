<template>
  <div v-if="visible" class="file-preview-modal" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <span class="file-name">{{ fileName }}</span>
        <div class="tools" v-if="isPdf">
            <button class="tool-btn" @click="prevPage" :disabled="page <= 1">Prev</button>
            <span class="page-info">{{ page }} / {{ pageCount }}</span>
            <button class="tool-btn" @click="nextPage" :disabled="page >= pageCount">Next</button>
        </div>
        <button class="close-btn" @click="close">&times;</button>
      </div>
      <div class="modal-body">
         <div v-if="loading" class="loading-container">
            <div class="spinner"></div>
            <p>{{ status }}</p>
         </div>
         <div v-else-if="isPdf" class="pdf-container">
             <VuePdfEmbed 
                ref="pdfRef"
                :source="fileUrl" 
                :page="page" 
                @loaded="handleLoaded"
             />
         </div>
         <img v-else-if="isImage" :src="fileUrl" alt="Preview" />
         <div v-else class="unsupported-msg">
            Preview non disponible pour ce type de fichier : {{ mimeType }} ({{ fileName }}) <br>
            <a :href="fileUrl" :download="fileName">Télécharger</a>
         </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import VuePdfEmbed from 'vue-pdf-embed'

// Essential for PDF.js in Vite
import * as pdfjsLib from 'pdfjs-dist';
import pdfWorker from 'pdfjs-dist/build/pdf.worker.mjs?url';

pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker;

const props = defineProps({
  visible: Boolean,
  fileUrl: String,
  fileName: String,
  mimeType: String,
  loading: Boolean,
  status: String
});

const emit = defineEmits(['close']);
const page = ref(1);
const pageCount = ref(1);
const pdfRef = ref(null);

const handleLoaded = (pdfDoc) => {
    if (pdfDoc && pdfDoc.numPages) {
        pageCount.value = pdfDoc.numPages;
    }
};

const nextPage = () => {
    if (page.value < pageCount.value) page.value++;
};
const prevPage = () => {
    if (page.value > 1) page.value--;
};

const close = () => {
  emit('close');
};

const isPdf = computed(() => {
    if (props.mimeType && props.mimeType.toLowerCase().includes('pdf')) return true;
    if (props.fileName && props.fileName.toLowerCase().endsWith('.pdf')) return true;
    return false;
});
const isImage = computed(() => props.mimeType && props.mimeType.startsWith('image/'));

watch(() => props.fileUrl, () => {
    page.value = 1;
    pageCount.value = 1;
});
</script>

<style scoped>
.file-preview-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.85); /* Overlay here */
}

/* Removed separate overlay div to use click.self on container */

.modal-content {
  position: relative;
  width: 55vw; /* Portrait style width */
  max-width: 900px;
  min-width: 350px;
  height: 95vh;
  background-color: #222;
  display: flex;
  flex-direction: column;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,0.5);
}

@media (max-width: 768px) {
    .modal-content {
        width: 95vw;
        height: 90vh;
    }
}

.modal-header {
  padding: 5px 15px;
  background-color: #f5f5f5;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #333;
  font-size: 0.9rem;
}

.tools {
    display: flex;
    gap: 8px;
    align-items: center;
}

.tool-btn {
    padding: 3px 8px;
    font-size: 0.85rem;
    cursor: pointer;
    background: #e0e0e0;
    border: none;
    border-radius: 4px;
}
.tool-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  color: #666;
}

.close-btn:hover {
    color: #000;
}

.modal-body {
  flex: 1;
  background-color: #525659;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: auto; /* Allow scrolling content */
  position: relative;
  padding: 10px;
}

.pdf-container {
    box-shadow: 0 0 10px rgba(0,0,0,0.5);
    max-width: 100%;
}

img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.unsupported-msg {
    color: white;
    text-align: center;
}
.unsupported-msg a {
    color: #4CAF50;
}

.loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    color: #ddd;
}
.spinner {
    border: 4px solid rgba(255, 255, 255, 0.1);
    width: 36px;
    height: 36px;
    border-radius: 50%;
    border-left-color: #4CAF50;
    animation: spin 1s linear infinite;
    margin-bottom: 10px;
}
@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}
</style>
