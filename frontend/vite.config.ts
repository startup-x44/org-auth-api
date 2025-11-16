import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    open: true,
  },
  build: {
    outDir: 'build',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks
          vendor: ['react', 'react-dom', 'react-router-dom'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu', '@radix-ui/react-tabs', '@radix-ui/react-toast', '@radix-ui/react-alert-dialog'],
          icons: ['lucide-react'],
          state: ['zustand'],
          utils: ['axios', 'framer-motion'],
          query: ['@tanstack/react-query'],
        },
      },
    },
    chunkSizeWarningLimit: 1000, // Increase warning limit to 1MB
  },
})
