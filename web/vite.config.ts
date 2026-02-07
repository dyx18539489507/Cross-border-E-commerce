import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3012,
    proxy: {
      '/api/v1/music/netease': {
        target: 'https://music.163.com',
        changeOrigin: true,
        secure: false,
        headers: {
          Referer: 'https://music.163.com',
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
        },
        rewrite: (path) => {
          if (path.startsWith('/api/v1/music/netease/search')) {
            return path.replace('/api/v1/music/netease/search', '/api/search/get')
          }
          if (path.startsWith('/api/v1/music/netease/song-url')) {
            return path.replace('/api/v1/music/netease/song-url', '/song/media/outer/url')
          }
          return path.replace('/api/v1/music/netease', '')
        }
      },
      '/netease': {
        target: 'https://music.163.com',
        changeOrigin: true,
        secure: false,
        headers: {
          Referer: 'https://music.163.com',
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
        },
        rewrite: (path) => path.replace(/^\/netease/, '')
      },
      '/api': {
        target: 'http://localhost:5678',
        changeOrigin: true
      }
    }
  }
})
