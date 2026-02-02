import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
    base: "/elibrary/",
    plugins: [react()],
    server: process.env.VITE_API_PROXY_TARGET
        ? {
              proxy: {
                  "/elibrary/api": {
                      target: process.env.VITE_API_PROXY_TARGET,
                      changeOrigin: true,
                      rewrite: (path) =>
                          path.replace(/^\/elibrary\/api/, ""),
                  },
              },
          }
        : undefined,
})
