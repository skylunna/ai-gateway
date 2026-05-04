import { useEffect, useState } from 'react';
import { Routes, Route, Navigate, NavLink, useLocation } from 'react-router-dom';
import { LayoutGrid, Activity, Shield, Settings as SettingsIcon, Search } from 'lucide-react';
import { Dashboard } from './pages/Dashboard';
import { TraceList } from './pages/TraceList';
import { TraceDetail } from './pages/TraceDetail';
import { Policies } from './pages/Policies';
import { Settings } from './pages/Settings';
import { fetchDashboardSummary } from './api/client';

const NAV = [
  { to: '/dashboard', label: 'Dashboard', Icon: LayoutGrid },
  { to: '/traces',    label: 'Traces',    Icon: Activity   },
  { to: '/policies',  label: 'Policies',  Icon: Shield     },
  { to: '/settings',  label: 'Settings',  Icon: SettingsIcon },
];

const PAGE_TITLE: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/traces':    'Traces',
  '/policies':  'Policies',
  '/settings':  'Settings',
};

function getTitle(pathname: string) {
  if (pathname.startsWith('/traces/')) return 'Trace Detail';
  return PAGE_TITLE[pathname] ?? 'luner';
}

export default function App() {
  const location  = useLocation();
  const title     = getTitle(location.pathname);
  const [count, setCount] = useState<number | null>(null);

  useEffect(() => {
    fetchDashboardSummary()
      .then(s => setCount(s.total_traces))
      .catch(() => {});
  }, []);

  return (
    <div className="flex h-screen overflow-hidden bg-surface-950">

      {/* ── Sidebar ─────────────────────────────────────────── */}
      <aside className="w-[140px] flex flex-col border-r border-surface-600 flex-shrink-0 bg-surface-950">

        {/* Logo */}
        <div className="flex items-center gap-2 px-4 py-[14px] border-b border-surface-600">
          <span className="text-amber-400 text-sm leading-none">🌙</span>
          <span className="font-bold text-white text-[13px] tracking-tight">luner</span>
          <span className="text-[9px] font-semibold bg-primary-900/50 text-primary-400
                           border border-primary-700/40 px-1 py-0.5 rounded ml-0.5 select-none">
            v0.5
          </span>
        </div>

        {/* Nav */}
        <nav className="flex-1 px-2.5 py-4">
          <p className="text-[10px] font-semibold text-slate-600 uppercase tracking-widest mb-2.5 px-1.5 select-none">
            Navigation
          </p>
          <div className="space-y-0.5">
            {NAV.map(({ to, label, Icon }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  `flex items-center gap-2 px-2 py-[7px] rounded-lg text-[12px] font-medium
                   transition-colors duration-100 ${
                    isActive
                      ? 'bg-primary-900/40 text-primary-400'
                      : 'text-slate-500 hover:text-slate-200 hover:bg-surface-800'
                  }`
                }
              >
                <Icon className="w-[14px] h-[14px] flex-shrink-0" />
                <span className="flex-1">{label}</span>
                {label === 'Traces' && count != null && (
                  <span className="text-[10px] tabular-nums text-slate-600">
                    {count.toLocaleString()}
                  </span>
                )}
              </NavLink>
            ))}
          </div>
        </nav>

        {/* Gateway status */}
        <div className="px-3.5 py-3 border-t border-surface-600">
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0 animate-pulse" />
            <span className="text-[11px] text-slate-500">Gateway</span>
            <span className="text-[11px] font-semibold text-emerald-400 ml-auto">online</span>
          </div>
        </div>
      </aside>

      {/* ── Main ────────────────────────────────────────────── */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">

        {/* Top bar */}
        <header className="h-12 flex items-center px-6 border-b border-surface-600 flex-shrink-0 bg-surface-950">
          <span className="text-[13px] font-semibold text-slate-300 mr-auto">{title}</span>

          {/* Search */}
          <div className="relative mr-3">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-[13px] h-[13px] text-slate-600 pointer-events-none" />
            <input
              type="text"
              placeholder="Quick search..."
              className="w-44 pl-8 pr-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                         text-[12px] text-slate-400 placeholder-slate-600
                         focus:ring-1 focus:ring-primary-600 focus:border-primary-600 outline-none
                         transition-all duration-150"
            />
          </div>

          {/* Avatar */}
          <div className="flex items-center gap-2 cursor-pointer select-none">
            <div className="w-7 h-7 rounded-full bg-primary-700 flex items-center justify-center
                            text-[11px] font-bold text-white">
              A
            </div>
            <span className="text-[12px] text-slate-400">admin</span>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto bg-surface-950">
          <div className="max-w-[1280px] mx-auto px-6 py-6">
            <Routes>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard"          element={<Dashboard />} />
              <Route path="/traces"             element={<TraceList />} />
              <Route path="/traces/:traceId"    element={<TraceDetail />} />
              <Route path="/policies"           element={<Policies />} />
              <Route path="/settings"           element={<Settings />} />
            </Routes>
          </div>
        </main>
      </div>
    </div>
  );
}
