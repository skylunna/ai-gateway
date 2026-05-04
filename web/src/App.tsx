import { Routes, Route, NavLink, Navigate } from 'react-router-dom';
import { TraceList } from './pages/TraceList';
import { TraceDetail } from './pages/TraceDetail';
import { Dashboard } from './pages/Dashboard';

export default function App() {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <nav className="bg-gray-900 text-white flex items-center gap-2 px-6 h-14 shadow-md flex-shrink-0">
        <span className="font-bold text-base tracking-tight text-white mr-auto select-none">
          ⬡ luner
        </span>
        <NavLink
          to="/dashboard"
          className={({ isActive }) =>
            `text-sm px-3 py-1.5 rounded-md transition-colors duration-150 ${
              isActive ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-800'
            }`
          }
        >
          Dashboard
        </NavLink>
        <NavLink
          to="/traces"
          className={({ isActive }) =>
            `text-sm px-3 py-1.5 rounded-md transition-colors duration-150 ${
              isActive ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-800'
            }`
          }
        >
          Traces
        </NavLink>
      </nav>

      <main className="flex-1 w-full max-w-6xl mx-auto px-6 py-6">
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
