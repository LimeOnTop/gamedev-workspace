import { Route, Routes } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import ThemePicker from './components/ThemePicker'
import Home from './pages/Home'
import NodePage from './pages/NodePage'
import AssetsPage from './pages/AssetsPage'
import { WorkspaceProvider } from './workspace'

export default function App() {
  return (
    <WorkspaceProvider>
      <div className="layout">
        <Sidebar />
        <main className="content">
          <div className="topbar">
            <ThemePicker />
          </div>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/node/:id" element={<NodePage />} />
            <Route path="/assets" element={<AssetsPage />} />
            <Route path="/assets/:category" element={<AssetsPage />} />
            <Route path="*" element={<Home />} />
          </Routes>
        </main>
      </div>
    </WorkspaceProvider>
  )
}
