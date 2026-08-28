import { Routes, Route, Navigate } from 'react-router-dom'
import { ToastProvider } from './store/useToast'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import ServerPage from './pages/ServerPage'
import ChannelsPage from './pages/ChannelsPage'
import KeysPage from './pages/KeysPage'

export default function App() {
  return (
    <ToastProvider>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/server" element={<ServerPage />} />
          <Route path="/channels" element={<ChannelsPage />} />
          <Route path="/keys" element={<KeysPage />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Route>
      </Routes>
    </ToastProvider>
  )
}
