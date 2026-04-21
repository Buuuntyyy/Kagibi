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
        manualChunks: (id) => {
          if (id.includes('libsodium-wrappers-sumo')) return 'libsodium'
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
