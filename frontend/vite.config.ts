import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const backend = process.env.BACKEND_URL ?? 'http://localhost:8081'

export default defineConfig({
  plugins: [react()],
  // Keep build output away from the SPA's /assets/* catalog routes.
  build: { assetsDir: '_app' },
  server: {
    port: 5173,
    proxy: {
      '/api': backend,
      '/uploads': backend,
    },
  },
})
