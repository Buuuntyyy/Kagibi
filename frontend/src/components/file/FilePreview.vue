<template>
  <div v-if="visible" class="file-preview-modal" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <span class="file-name">{{ fileName }}</span>
        <div class="tools" v-if="isPdf || isImage">
            <template v-if="isPdf">
                <button class="tool-btn" @click="prevPage" :disabled="page <= 1">Prev</button>
                <span class="page-info">{{ page }} / {{ pageCount }}</span>
                <button class="tool-btn" @click="nextPage" :disabled="page >= pageCount">Next</button>
                <div class="separator"></div>
            </template>
            <button class="tool-btn" @click="zoomOut">-</button>
            <span class="page-info">{{ Math.round(scale * 100) }}%</span>
            <button class="tool-btn" @click="zoomIn">+</button>
        </div>
        <button class="close-btn" @click="close">&times;</button>
      </div>
      <div class="modal-body">
         <!-- Phase 1: Downloading/Preparing (Server/Decrypt) -->
         <div v-if="loading" class="loading-container">
            <div class="spinner"></div>
            <p>{{ status || 'Chargement...' }}</p>
         </div>

         <!-- Phase 2: Display -->
         <template v-else>
            <!-- PDF Viewer -->
            <div v-if="isPdf" class="pdf-wrapper" :style="{ width: '100%' }">
                 <div v-if="isRendering" class="loading-container" style="position: absolute; inset: 0; background: #525659; z-index: 10;">
                    <div class="spinner"></div>
                    <p>Rendu du PDF...</p>
                 </div>
                 <div class="pdf-container" :style="{ width: scale * 100 + '%' }">
                     <VuePdfEmbed 
                        :key="scale"
                        ref="pdfRef"
                        :source="pdfSource" 
                        :page="page" 
                        @loaded="handleLoaded"
                        @loading-failed="handleError"
                     />
                 </div>
            </div>

            <!-- Image Viewer -->
            <div v-else-if="isImage" style="display: flex; justify-content: center; min-height: 100%;">
                <img :src="fileUrl" alt="Preview" :style="{ width: scale * 100 + '%', maxWidth: 'none', maxHeight: 'none' }" />
            </div>

            <!-- Unsupported -->
            <div v-else class="unsupported-msg">
                Preview non disponible pour ce type de fichier : {{ mimeType }} ({{ fileName }}) <br>
                <a :href="fileUrl" :download="fileName">Télécharger</a>
            </div>
         </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch, onMounted } from 'vue';
import VuePdfEmbed from 'vue-pdf-embed'

// Essential for PDF.js in Vite
import * as pdfjsLib from 'pdfjs-dist';
// V5/V4 worker import
import pdfWorker from 'pdfjs-dist/build/pdf.worker.mjs?url';

if (typeof window !== 'undefined' && 'Worker' in window) {
    pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker;
}

const props = defineProps({
  visible: Boolean,
  fileUrl: String,
  fileName: String,
  mimeType: String,
  loading: Boolean,
  status: String
});

const emit = defineEmits(['close']);

const isPdf = computed(() => {
    // 1. Explicit PDF MimeType
    if (props.mimeType && props.mimeType.toLowerCase().includes('pdf')) return true;
    
    // 2. Explicit Image MimeType -> Not PDF (even if filename ends in .pdf, it might be a server-side preview)
    if (props.mimeType && props.mimeType.toLowerCase().startsWith('image/')) return false;

    // 3. Fallback to extension
    if (props.fileName && props.fileName.toLowerCase().endsWith('.pdf')) return true;
    return false;
});

const isImage = computed(() => {
    if (props.mimeType && props.mimeType.startsWith('image/')) return true;
    if (props.fileName) {
        const lowerName = props.fileName.toLowerCase();
        return lowerName.endsWith('.jpg') || 
               lowerName.endsWith('.jpeg') || 
               lowerName.endsWith('.png') || 
               lowerName.endsWith('.gif') || 
               lowerName.endsWith('.webp') || 
               lowerName.endsWith('.bmp') ||
               lowerName.endsWith('.svg');
    }
    return false;
});

// Removed internal 'loading' state mutation, use local
// Actually props.loading is passed from store. We should not mutate it directly but the store handles it?
// The store sets loading=false when download completes.
// But the content rendering (PDF) is async too.
const isRendering = ref(true);

watch(() => props.loading, (newVal) => {
    isRendering.value = newVal; // Sync with prop
    // But PDF rendering starts AFTER prop loading is false (when URL is ready)
    if (!newVal && props.fileUrl && isPdf.value) {
        isRendering.value = true; // Start PDF rendering wait
    } else if (!newVal) {
        isRendering.value = false;
    }
});
const page = ref(1);
const pageCount = ref(1);
const pdfRef = ref(null);
const scale = ref(1.0);
const pdfSource = ref(null); // Document Proxy or URL

const handleLoaded = (pdfDoc) => {
    console.log("PDF Loaded successfully", pdfDoc);
    isRendering.value = false;
    if (pdfDoc && pdfDoc.numPages) {
        pageCount.value = pdfDoc.numPages;
    }
};

const handleError = (error) => {
    console.error("PDF Preview Error (VuePdfEmbed):", error);
    isRendering.value = false;
}

// Manually load PDF to debug and ensure control
const loadPdf = async (url) => {
    if (!url) {
        pdfSource.value = null;
        return;
    }
    console.log("Loading PDF from URL:", url);
    isRendering.value = true;
    
    try {
        // Option A: Pass URL directly (let VuePdfEmbed handle it)
        // pdfSource.value = url;
        
        // Option B: Load manually
        const loadingTask = pdfjsLib.getDocument({
             url: url,
             cMapUrl: 'https://unpkg.com/pdfjs-dist@4.10.0/cmaps/', // Use CDN for cmaps to avoid local 404s
             cMapPacked: true,
        });
        
        const doc = await loadingTask.promise;
        console.log("PDF Document loaded manually:", doc.numPages, "pages");
        pdfSource.value = doc;
        // isRendering will be set to false by handleLoaded when component renders it
        
    } catch (e) {
        console.error("Manual PDF Load Failed:", e);
        handleError(e);
        pdfSource.value = null; // Reset
    }
};

watch(() => props.fileUrl, (newUrl) => {
    page.value = 1;
    pageCount.value = 1;
    scale.value = 1.0;
    if (newUrl && isPdf.value) {
        loadPdf(newUrl);
    }
}, { immediate: true });

const nextPage = () => {
    if (page.value < pageCount.value) page.value++;
};

const prevPage = () => {
    if (page.value > 1) page.value--;
};

const zoomIn = () => {
    scale.value = Math.min(scale.value + 0.1, 5.0); // Max 500%
};

const zoomOut = () => {
    scale.value = Math.max(scale.value - 0.1, 0.1); // Min 10%
};

const close = () => {
  emit('close');
};


watch(() => props.fileUrl, () => {
    page.value = 1;
    pageCount.value = 1;
    scale.value = 1.0;
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

.separator {
    width: 1px;
    height: 20px;
    background-color: #ccc;
    margin: 0 10px;
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
  overflow: auto; /* Allow scrolling content */
  position: relative;
  padding: 10px;
  display: flex; /* Use flex for easy centering with margin: auto */
}

.pdf-container, img, .loading-container, .unsupported-msg {
    margin: auto; /* Magic centering: centers when small, allows scroll when big */
}

.pdf-wrapper {
    margin: auto;
    position: relative; 
    min-height: 200px;
    display: flex;
    justify-content: center;
}

.pdf-container {
    box-shadow: 0 0 10px rgba(0,0,0,0.5);
}

img {
  display: block; /* Important for margin: auto to work if not flex child */
  max-width: none; /* Allow zoom */
  max-height: none; /* Allow zoom */
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
