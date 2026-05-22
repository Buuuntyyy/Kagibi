import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  optimizeDeps: {
    // Exclude libsodium from pre-bundling: Rolldown would hit its Node.js
    // require('crypto') and externalize it, breaking the optimization step.
    // libsodium auto-detects the environment and uses its WASM build in browsers.
    exclude: ['libsodium-sumo', 'libsodium-wrappers-sumo'],
  },
  build: {
    target: 'esnext',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('libsodium')) return 'libsodium'
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
