import { useEffect, useState } from 'react';
import { Routes, Route, Navigate, NavLink, useLocation } from 'react-router-dom';
import { TraceList } from './pages/TraceList';
import { TraceDetail } from './pages/TraceDetail';
import { Dashboard } from './pages/Dashboard';

const PAGE_INFO: Record<string, { title: string }> = {
  '/dashboard': { title: 'Dashboard' },
  '/traces':    { title: 'Traces' },
};

function getTitle(pathname: string) {
  if (pathname.startsWith('/traces/')) return 'Trace Detail';
  return PAGE_INFO[pathname]?.title ?? 'luner';
}

function UtcClock() {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 30_000);
    return () => clearInterval(id);
  }, []);
  const MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  const m = MONTHS[now.getUTCMonth()];
  const d = now.getUTCDate();
  const y = now.getUTCFullYear();
  const hh = String(now.getUTCHours()).padStart(2, '0');
  const mm = String(now.getUTCMinutes()).padStart(2, '0');
  return (
    <span className="text-slate-500 text-xs tabular-nums select-none">
      {m} {d}, {y} · {hh}:{mm} UTC
    </span>
  );
}

export default function App() {
  const location = useLocation();
  const title = getTitle(location.pathname);

  return (
    <div className="min-h-screen flex flex-col bg-surface-950">
      {/* Navbar */}
      <nav className="flex items-center px-6 h-[60px] border-b border-surface-600 flex-shrink-0 bg-surface-950">
        {/* Brand */}
        <div className="flex items-center gap-2">
          <span className="text-amber-400 text-base leading-none select-none">🌙</span>
          <span className="font-bold text-white text-[15px] tracking-tight">luner</span>
          <span className="text-[10px] font-semibold bg-primary-900/50 text-primary-400
                           border border-primary-700/40 px-1.5 py-0.5 rounded select-none">
            v0.5
          </span>
        </div>

        {/* Divider + page title */}
        <div className="w-px h-5 bg-surface-500 mx-4" />
        <div className="flex flex-col leading-tight mr-auto">
          <span className="text-slate-200 font-semibold text-sm">{title}</span>
          <span className="text-slate-600 text-[11px]">LLM API Gateway</span>
        </div>

        {/* Nav links */}
        <div className="flex items-center gap-1 mr-5">
          {[
            { to: '/dashboard', label: 'Dashboard' },
            { to: '/traces',    label: 'Traces' },
          ].map(({ to, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `text-xs px-3 py-1.5 rounded-md transition-colors duration-150 ${
                  isActive
                    ? 'bg-surface-700 text-white border border-surface-500'
                    : 'text-slate-500 hover:text-slate-200 hover:bg-surface-800'
                }`
              }
            >
              {label}
            </NavLink>
          ))}
        </div>

        <UtcClock />
      </nav>

      {/* Page content */}
      <main className="flex-1 w-full max-w-[1280px] mx-auto px-6 py-6">
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
