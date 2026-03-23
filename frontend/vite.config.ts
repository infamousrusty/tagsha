// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        advancedChunks: {
          groups: [
            { name: 'vendor', test: new RegExp('node_modules/react') },
          ],
        },
      },
    },
  },
  test: {
    environment: 'jsdom',
    passWithNoTests: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
})