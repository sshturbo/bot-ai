import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import MessagePage from './screens/MessagePage'
import '@/global.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/message/:hash" element={<MessagePage />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>
)
