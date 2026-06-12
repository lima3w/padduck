import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    port: 3000,
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('/react-router') || id.includes('/@remix-run')) return 'vendor-router'
            if (id.includes('/react') || id.includes('/scheduler')) return 'vendor-react'
            if (id.includes('/recharts') || id.includes('/d3-')) return 'vendor-charts'
            if (id.includes('/cytoscape')) return 'vendor-topology'
            if (id.includes('/axios')) return 'vendor-http'
            return 'vendor'
          }

          if (id.includes('/src/components/') || id.includes('/src/hooks/') || id.includes('/src/api/')) {
            return 'app-shared'
          }
        },
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.js',
    exclude: ['node_modules/**', 'e2e/**'],
    coverage: {
      provider: 'v8',
      // Regression floor, set just below current coverage of test-imported
      // files (38.7/36.5/28.2/21.2 as of the v1.31.25 api-module split). The
      // denominator moves when imports restructure or tests pull in large
      // untested pages, so recalibrate then; otherwise ratchet up as
      // coverage grows.
      thresholds: {
        lines: 36,
        statements: 34,
        branches: 26,
        functions: 19,
      },
    },
  },
})
