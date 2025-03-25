import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from "path"

// Obtém a URL da API do ambiente ou usa o padrão para desenvolvimento
const apiUrl = process.env.VITE_API_URL || 'http://localhost:8443'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: apiUrl,
        changeOrigin: true
      }
    },
    historyApiFallback: true
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // Define variáveis de ambiente que estarão disponíveis no cliente
  define: {
    'process.env.VITE_API_URL': JSON.stringify(apiUrl)
  }
})
