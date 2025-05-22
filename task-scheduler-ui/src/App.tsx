import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import BasicLayout from './layouts/BasicLayout'
import DashboardPage from './pages/Dashboard'
import TasksPage from './pages/Tasks'
import NodesPage from './pages/Nodes'
import LogsPage from './pages/Logs'
import SettingsPage from './pages/Settings'
import ErrorBoundary from './components/ErrorBoundary'
import { NotificationProvider } from './components/GlobalNotification'
import './App.css'

function App() {
  return (
    <ErrorBoundary>
      <NotificationProvider />
      <Router>
        <Routes>
          <Route path="/" element={<BasicLayout><Navigate to="/dashboard" replace /></BasicLayout>} />
          <Route path="/dashboard" element={<BasicLayout><DashboardPage /></BasicLayout>} />
          <Route path="/tasks" element={<BasicLayout><TasksPage /></BasicLayout>} />
          <Route path="/nodes" element={<BasicLayout><NodesPage /></BasicLayout>} />
          <Route path="/logs" element={<BasicLayout><LogsPage /></BasicLayout>} />
          <Route path="/settings" element={<BasicLayout><SettingsPage /></BasicLayout>} />
          <Route path="*" element={<BasicLayout><Navigate to="/dashboard" replace /></BasicLayout>} />
        </Routes>
      </Router>
    </ErrorBoundary>
  )
}

export default App