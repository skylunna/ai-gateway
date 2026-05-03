import { Routes, Route, NavLink, Navigate } from 'react-router-dom';
import { TraceList } from './pages/TraceList';
import { TraceDetail } from './pages/TraceDetail';
import { Dashboard } from './pages/Dashboard';

export default function App() {
  return (
    <div className="app">
      <nav className="navbar">
        <span className="brand">⬡ luner</span>
        <NavLink to="/dashboard" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          Dashboard
        </NavLink>
        <NavLink to="/traces" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          Traces
        </NavLink>
      </nav>
      <main className="content">
        <Routes>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/traces" element={<TraceList />} />
          <Route path="/traces/:traceId" element={<TraceDetail />} />
        </Routes>
      </main>
    </div>
  );
}
