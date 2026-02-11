import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 9000,
    host: true, // Expose to network if needed
    allowedHosts: [
      'prescholastic-hurtlingly-latia.ngrok-free.dev'
    ],
    proxy: {
      '/api': {
        target: 'http://localhost:5039',
        changeOrigin: true,
        secure: false,
      }
    }
  }
})
