import { Route, Routes } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import Home from './pages/Home'
import NodePage from './pages/NodePage'
import { WorkspaceProvider } from './workspace'

export default function App() {
  return (
    <WorkspaceProvider>
      <div className="layout">
        <Sidebar />
        <main className="content">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/node/:id" element={<NodePage />} />
            <Route path="*" element={<Home />} />
          </Routes>
        </main>
      </div>
    </WorkspaceProvider>
  )
}
