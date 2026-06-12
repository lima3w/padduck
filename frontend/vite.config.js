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
    coverage: {
      provider: 'v8',
      // Regression floor, set just below current coverage of test-imported
      // files (49.7/41.8/33.9/17.4 as of v1.31.25). Ratchet up as tests grow.
      thresholds: {
        lines: 45,
        statements: 38,
        branches: 30,
        functions: 15,
      },
    },
  },
})
