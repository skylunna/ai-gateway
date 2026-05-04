import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Search, RefreshCw } from 'lucide-react';
import { fetchTraces } from '../api/client';
import type { TraceItem } from '../types/trace';
import { format } from 'date-fns';

function modelBadgeClass(model: string): string {
  if (model.toLowerCase().startsWith('claude'))   return 'bg-violet-900/40 text-violet-300 border border-violet-700/30';
  if (model.toLowerCase().includes('mini'))        return 'bg-cyan-900/40  text-cyan-300  border border-cyan-700/30';
  if (model.toLowerCase().startsWith('gpt'))       return 'bg-blue-900/40  text-blue-300  border border-blue-700/30';
  return 'bg-slate-700/40 text-slate-400 border border-slate-600/30';
}

const TH = 'px-4 py-3 text-left text-[11px] font-semibold text-slate-500 uppercase tracking-wider bg-surface-900 border-b border-surface-500';
const TD = 'px-4 py-3 text-sm text-slate-300';

export function TraceList() {
  const [traces,      setTraces]      = useState<TraceItem[]>([]);
  const [total,       setTotal]       = useState(0);
  const [page,        setPage]        = useState(1);
  const [agentFilter, setAgentFilter] = useState('');
  const [userFilter,  setUserFilter]  = useState('');
  const [loading,     setLoading]     = useState(true);
  const [error,       setError]       = useState('');

  const load = useCallback(() => {
    setLoading(true);
    setError('');
    fetchTraces({ page, page_size: 20, agent_name: agentFilter || undefined, user_id: userFilter || undefined })
      .then(data => { setTraces(data.traces ?? []); setTotal(data.total_count ?? 0); })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [page, agentFilter, userFilter]);

  useEffect(() => { load(); }, [load]);

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3 mb-5">
        <div className="flex-1">
          <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
            {total} TRACES · LAST 24H
          </p>
        </div>
        <button onClick={load} className="btn-secondary flex items-center gap-2 self-start sm:self-auto">
          <RefreshCw className="w-3.5 h-3.5" />
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-4 flex-wrap">
        {[
          { placeholder: 'Filter by agent…', value: agentFilter, onChange: (v: string) => { setAgentFilter(v); setPage(1); } },
          { placeholder: 'Filter by user…',  value: userFilter,  onChange: (v: string) => { setUserFilter(v);  setPage(1); } },
        ].map(({ placeholder, value, onChange }) => (
          <div key={placeholder} className="relative flex-1 min-w-[160px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-600 pointer-events-none" />
            <input
              type="text"
              placeholder={placeholder}
              value={value}
              onChange={e => onChange(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-surface-800 border border-surface-500 rounded-lg text-sm
                         text-slate-300 placeholder-slate-600
                         focus:ring-1 focus:ring-primary-600 focus:border-primary-600 outline-none
                         transition-all duration-150"
            />
          </div>
        ))}
      </div>

      {error && (
        <div className="text-red-400 bg-red-900/20 border border-red-800/40 rounded-lg px-4 py-3 mb-4 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-600 text-sm">Loading…</div>
      ) : (
        <>
          <div className="table-container">
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  <th className={TH}>Trace ID</th>
                  <th className={TH}>Agent</th>
                  <th className={TH}>User</th>
                  <th className={TH}>Model</th>
                  <th className={TH}>Spans</th>
                  <th className={TH}>Cost</th>
                  <th className={TH}>Duration</th>
                  <th className={TH}>Status</th>
                  <th className={TH}>Started</th>
                </tr>
              </thead>
              <tbody>
                {traces.length === 0 && (
                  <tr>
                    <td colSpan={9} className="text-center py-16 text-slate-600 text-sm">
                      No traces found
                    </td>
                  </tr>
                )}
                {traces.map(t => (
                  <tr key={t.trace_id} className="table-row">
                    <td className={TD}>
                      <Link
                        to={`/traces/${t.trace_id}`}
                        className="text-primary-400 hover:text-primary-300 hover:underline font-mono text-xs"
                      >
                        {t.trace_id.slice(0, 8)}…
                      </Link>
                    </td>
                    <td className={TD}>{t.agent_name ?? <span className="text-slate-600">—</span>}</td>
                    <td className={`${TD} text-slate-400`}>{t.user_id ?? <span className="text-slate-600">—</span>}</td>
                    <td className={TD}>
                      {t.model ? (
                        <span className={`inline-block px-2 py-0.5 rounded text-[11px] font-medium ${modelBadgeClass(t.model)}`}>
                          {t.model}
                        </span>
                      ) : <span className="text-slate-600">—</span>}
                    </td>
                    <td className={`${TD} text-slate-400`}>{t.span_count}</td>
                    <td className={`${TD} text-primary-400 font-semibold tabular-nums`}>
                      ${t.total_cost_usd.toFixed(4)}
                    </td>
                    <td className={`${TD} text-slate-400 tabular-nums`}>
                      {t.duration_ms > 0 ? `${t.duration_ms}ms` : '—'}
                    </td>
                    <td className={TD}>
                      <span className={`badge ${
                        t.status === 'success' ? 'badge-success' :
                        t.status === 'error'   ? 'badge-error'   : 'badge-warning'
                      }`}>
                        {t.status}
                      </span>
                    </td>
                    <td className={`${TD} text-slate-500 whitespace-nowrap tabular-nums`}>
                      {format(new Date(t.start_time), 'MMM d HH:mm:ss')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between mt-4 text-xs text-slate-500">
            <span>{total} total</span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
                className="px-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                           hover:bg-surface-700 disabled:opacity-30 disabled:cursor-not-allowed
                           transition-colors text-xs text-slate-400"
              >
                ← Prev
              </button>
              <span className="px-2 font-medium text-slate-400">Page {page}</span>
              <button
                disabled={traces.length < 20}
                onClick={() => setPage(p => p + 1)}
                className="px-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                           hover:bg-surface-700 disabled:opacity-30 disabled:cursor-not-allowed
                           transition-colors text-xs text-slate-400"
              >
                Next →
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
