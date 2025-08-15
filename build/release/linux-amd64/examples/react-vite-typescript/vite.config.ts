import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'fs'
import path from 'path'

export default defineConfig(() => {
  // Read NSM configuration if available
  let port = {{.Port}}
  let host = '127.0.0.1'
  
  const nsmPortsFile = path.resolve('.nsm-ports.json')
  if (fs.existsSync(nsmPortsFile)) {
    try {
      const config = JSON.parse(fs.readFileSync(nsmPortsFile, 'utf8'))
      port = config.http || port
      host = config.host || host
      console.log(`🔧 NSM: Using HTTP port ${port}`)
    } catch (e) {
      console.warn('NSM: Failed to read configuration:', e.message)
    }
  }

  return {
    plugins: [react()],
    server: {
      host,
      port,
      strictPort: true,
      open: false, // NSM handles browser opening
      cors: true,
    },
    build: {
      outDir: 'dist',
      sourcemap: true,
    },
    // NSM environment variables
    define: {
      'import.meta.env.NSM_ENABLED': JSON.stringify(!!process.env.NSM_ENABLED),
      'import.meta.env.NSM_DOMAIN': JSON.stringify('{{.Domain}}'),
    },
  }
})
