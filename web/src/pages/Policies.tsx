import { useEffect, useState } from 'react';
import { RefreshCw, Plus } from 'lucide-react';
import { fetchPolicies } from '../api/client';
import type { Policy } from '../types/trace';

const ACTION_BADGE: Record<string, string> = {
  block:     'bg-red-900/50    text-red-300    border border-red-700/40',
  alert:     'bg-amber-900/50  text-amber-300  border border-amber-700/40',
  downgrade: 'bg-blue-900/50   text-blue-300   border border-blue-700/40',
};

const ACTION_LABEL: Record<string, string> = {
  block:     'Block',
  alert:     'Alert',
  downgrade: 'Downgrade',
};

export function Policies() {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading,  setLoading]  = useState(true);
  const [error,    setError]    = useState('');

  const load = () => {
    setLoading(true);
    setError('');
    fetchPolicies()
      .then(setPolicies)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  return (
    <div className="animate-fade-in">
      {/* Page header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Policies</h1>
          <p className="text-xs text-slate-500 mt-0.5">Gateway routing, caching and rate-limit rules</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={load} className="btn-secondary flex items-center gap-1.5 py-1.5 text-xs">
            <RefreshCw className="w-3.5 h-3.5" />
            Refresh
          </button>
          <button className="btn-primary flex items-center gap-1.5 py-1.5 text-xs">
            <Plus className="w-3.5 h-3.5" />
            New Policy
          </button>
        </div>
      </div>

      {error && (
        <div className="text-red-400 bg-red-900/20 border border-red-800/40 rounded-lg px-4 py-3 mb-4 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-600 text-sm">Loading…</div>
      ) : (
        <div className="card overflow-hidden">
          {policies.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-600">
              <Shield className="w-8 h-8 mb-3 opacity-30" />
              <p className="text-sm">No policies defined</p>
              <p className="text-xs mt-1 text-slate-700">Create a policy to control LLM request behaviour</p>
            </div>
          ) : (
            policies.map((p, i) => (
              <div
                key={p.id}
                className={`flex items-center gap-4 px-5 py-3.5 transition-colors hover:bg-surface-750
                            ${i < policies.length - 1 ? 'border-b border-surface-600' : ''}`}
              >
                {/* Action badge */}
                <span className={`inline-block px-2 py-0.5 rounded text-[11px] font-bold
                                  uppercase tracking-wide flex-shrink-0 ${ACTION_BADGE[p.action] ?? ACTION_BADGE.block}`}>
                  {ACTION_LABEL[p.action] ?? p.action}
                </span>

                {/* Name */}
                <span className="text-sm font-semibold text-slate-200 flex-shrink-0">{p.name}</span>

                {/* Description */}
                {p.description && (
                  <span className="text-xs text-slate-500 flex-shrink-0">{p.description}</span>
                )}

                {/* Expression preview */}
                <span className="text-xs text-slate-600 font-mono truncate flex-1 min-w-0">
                  {p.expression}
                </span>

                {/* Priority */}
                {p.priority > 0 && (
                  <span className="text-[11px] text-slate-600 flex-shrink-0">
                    priority {p.priority}
                  </span>
                )}

                {/* Status */}
                <span className={`badge flex-shrink-0 ${p.enabled ? 'badge-success' : 'badge-warning'}`}>
                  {p.enabled ? 'active' : 'inactive'}
                </span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

// used in empty state
function Shield({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round"
        d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.955 11.955 0 003 10.5c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.249-8.25-3.286z" />
    </svg>
  );
}
