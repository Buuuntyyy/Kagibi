import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  optimizeDeps: {
    esbuildOptions: {
      target: 'esnext',
    },
  },
  build: {
    target: 'esnext',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'libsodium': ['libsodium-wrappers-sumo'],
        },
      },
    },
  },
  server: {
    sourcemap: false,
  },
  define: {
    'process.env.DEBUG': false,
  },
})
