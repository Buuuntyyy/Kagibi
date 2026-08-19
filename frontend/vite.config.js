import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'node:fs'
import path from 'node:path'

// Mirrors the repo-root changelog files (single source of truth, one level above
// frontend/) into public/ so they ship as static assets and ChangelogModal.vue can
// fetch() them at runtime instead of duplicating their content in JS. Runs on both
// `vite dev` and `vite build` so the copy is always fresh — no separate release step.
function syncChangelogPlugin() {
  const files = ['CHANGELOG.md', 'CHANGELOG.en.md']
  const sync = () => {
    for (const name of files) {
      const src = path.resolve(__dirname, '..', name)
      if (!fs.existsSync(src)) continue
      const dest = path.resolve(__dirname, 'public', name)
      fs.mkdirSync(path.dirname(dest), { recursive: true })
      fs.copyFileSync(src, dest)
    }
  }
  return {
    name: 'sync-changelog',
    configResolved: sync,
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), syncChangelogPlugin()],
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
