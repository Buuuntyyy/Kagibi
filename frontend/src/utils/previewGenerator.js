import * as pdfjsLib from 'pdfjs-dist';
import pdfWorker from 'pdfjs-dist/build/pdf.worker.mjs?url';

// Configure Worker
pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker;

const PREVIEW_MAX_WIDTH = 1024;
const PREVIEW_QUALITY = 0.7;

export async function generatePreview(file) {
    try {
        if (file.type.startsWith('image/')) {
            return await generateImagePreview(file);
        }
        if (file.type === 'application/pdf') {
            return await generatePdfPreview(file);
        }
    } catch (e) {
        console.warn("Preview generation failed", e);
    }
    return null;
}

function generateImagePreview(file) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        const url = URL.createObjectURL(file);
        img.onload = () => {
            URL.revokeObjectURL(url);
            const canvas = document.createElement('canvas');
            let width = img.width;
            let height = img.height;

            if (width > PREVIEW_MAX_WIDTH) {
                height = Math.round(height * (PREVIEW_MAX_WIDTH / width));
                width = PREVIEW_MAX_WIDTH;
            }

            canvas.width = width;
            canvas.height = height;
            const ctx = canvas.getContext('2d');
            ctx.drawImage(img, 0, 0, width, height);

            canvas.toBlob((blob) => {
                if (blob) resolve(blob);
                else reject(new Error("Canvas toBlob failed"));
            }, 'image/jpeg', PREVIEW_QUALITY);
        };
        img.onerror = (e) => {
            URL.revokeObjectURL(url);
            reject(new Error(e.message));
        };
        img.src = url;
    });
}

async function generatePdfPreview(file) {
    const arrayBuffer = await file.arrayBuffer();
    // Use standard font map to avoid warnings/missing text
    const loadingTask = pdfjsLib.getDocument({ 
        data: arrayBuffer,
        cMapUrl: 'https://unpkg.com/pdfjs-dist@4.10.0/cmaps/',
        cMapPacked: true,
    });
    const pdf = await loadingTask.promise;
    const page = await pdf.getPage(1);
    
    const scale = 1.0;
    const viewport = page.getViewport({ scale });
    
    // Calculate scale to fit width
    let finalScale = 1.0;
    if (viewport.width > PREVIEW_MAX_WIDTH) {
        finalScale = PREVIEW_MAX_WIDTH / viewport.width;
    }
    const scaledViewport = page.getViewport({ scale: finalScale });

    const canvas = document.createElement('canvas');
    canvas.width = scaledViewport.width;
    canvas.height = scaledViewport.height;
    const ctx = canvas.getContext('2d');

    const renderContext = {
        canvasContext: ctx,
        viewport: scaledViewport
    };

    await page.render(renderContext).promise;

    return new Promise((resolve) => {
        canvas.toBlob((blob) => {
            resolve(blob);
        }, 'image/jpeg', PREVIEW_QUALITY);
    });
}
