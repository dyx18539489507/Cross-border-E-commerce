import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendTarget = env.VITE_DEV_PROXY_TARGET || 'http://127.0.0.1:5678'

  return {
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
        target: backendTarget,
        changeOrigin: true
      },
      '/static': {
        target: backendTarget,
        changeOrigin: true
      },
      '/health': {
        target: backendTarget,
        changeOrigin: true
      }
    }
    }
  }
})
