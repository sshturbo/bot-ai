import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import MessagePage from './screens/MessagePage'
import '@/global.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/chat/new" replace />} />
        <Route path="/chat/new" element={<MessagePage />} />
        <Route path="/chat/:hash" element={<MessagePage />} />
        {/* Mantém a rota antiga por compatibilidade */}
        <Route path="/message/:hash" element={<MessagePage />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>
)
