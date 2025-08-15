import { useState } from 'react'
import './App.css'

function App() {
  const [count, setCount] = useState(0)
  
  // NSM environment info
  const nsmEnabled = import.meta.env.NSM_ENABLED === 'true'
  const nsmDomain = import.meta.env.NSM_DOMAIN || 'localhost'

  return (
    <div className="App">
      <div className="header">
        <h1>🚀 NSM + React + Vite</h1>
        <p>{{.Description}}</p>
      </div>

      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>

      {nsmEnabled && (
        <div className="nsm-info">
          <h3>🔧 NSM Development Environment</h3>
          <ul>
            <li>Domain: <strong>{nsmDomain}</strong></li>
            <li>HTTPS: <strong>Enabled</strong></li>
            <li>Hot Reload: <strong>Active</strong></li>
          </ul>
          <p>
            Your app is running in a professional development environment with:
            clean URLs, automatic HTTPS, and custom domain resolution.
          </p>
        </div>
      )}

      <div className="links">
        <a href="https://vitejs.dev" target="_blank">
          Learn Vite
        </a>
        {' | '}
        <a href="https://react.dev" target="_blank">
          Learn React
        </a>
        {' | '}
        <a href="https://github.com/kashifsb/nsm" target="_blank">
          NSM Documentation
        </a>
      </div>
    </div>
  )
}

export default App
