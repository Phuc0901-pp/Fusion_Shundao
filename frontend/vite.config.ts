import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { compression } from 'vite-plugin-compression2';
import { VitePWA } from 'vite-plugin-pwa';
import path from 'path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // Gzip + Brotli compression for assets
    compression({ algorithms: ['gzip'], exclude: [/\.(png|jpg|jpeg|gif|webp|svg|ico)$/] }),
    compression({ algorithms: ['brotliCompress'], exclude: [/\.(png|jpg|jpeg|gif|webp|svg|ico)$/] }),
    // ─── PWA ────────────────────────────────────────────────────────────────
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'pwa-192.png', 'pwa-512.png'],
      manifest: {
        name: 'Fusion Shundao Dashboard',
        short_name: 'Shundao',
        description: 'Hệ thống giám sát điện mặt trời Shundao',
        theme_color: '#0f172a',
        background_color: '#0f172a',
        display: 'standalone',
        orientation: 'any',
        start_url: '/',
        icons: [
          { src: 'pwa-192.png', sizes: '192x192', type: 'image/png' },
          { src: 'pwa-512.png', sizes: '512x512', type: 'image/png' },
          { src: 'pwa-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],
      },
      workbox: {
        // Cache static JS/CSS/HTML indefinitely (they have content-hash filenames)
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        // Network-first for API calls - ensures fresh data
        runtimeCaching: [
          {
            urlPattern: /^http:\/\/localhost:5039\/api\/.*/,
            handler: 'NetworkOnly',
          },
        ],
      },
    }),
  ],
  server: {
    port: 9000,
    host: true,
    allowedHosts: ['prescholastic-hurtlingly-latia.ngrok-free.dev'],
    proxy: {
      '/api': {
        target: 'http://localhost:5039',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-query': ['@tanstack/react-query'],
          'vendor-charts': ['recharts'],
          'vendor-icons': ['lucide-react'],
          'vendor-http': ['axios', 'sonner'],
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});

