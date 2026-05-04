import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Search, RefreshCw } from 'lucide-react';
import { fetchTraces } from '../api/client';
import type { TraceItem } from '../types/trace';
import { format } from 'date-fns';

type StatusFilter = 'all' | 'success' | 'error' | 'timeout';

function modelBadgeClass(model: string): string {
  const m = model.toLowerCase();
  if (m.startsWith('claude')) return 'bg-violet-900/40 text-violet-300 border border-violet-700/30';
  if (m.includes('mini'))     return 'bg-cyan-900/40  text-cyan-300  border border-cyan-700/30';
  if (m.startsWith('gpt'))    return 'bg-blue-900/40  text-blue-300  border border-blue-700/30';
  return 'bg-slate-700/40 text-slate-400 border border-slate-600/30';
}

const TH = 'px-4 py-3 text-left text-[11px] font-semibold text-slate-500 uppercase tracking-wider bg-surface-900 border-b border-surface-500';
const TD = 'px-4 py-3 text-sm';

export function TraceList() {
  const [traces,      setTraces]      = useState<TraceItem[]>([]);
  const [total,       setTotal]       = useState(0);
  const [page,        setPage]        = useState(1);
  const [agentFilter, setAgentFilter] = useState('');
  const [userFilter,  setUserFilter]  = useState('');
  const [status,      setStatus]      = useState<StatusFilter>('all');
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

  const visible = status === 'all'
    ? traces
    : traces.filter(t => t.status === status);

  const totalPages = Math.max(1, Math.ceil(total / 20));

  return (
    <div className="animate-fade-in">
      {/* Page header */}
      <div className="flex items-start justify-between mb-5">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Traces</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            {total} traces · page {page} of {totalPages}
          </p>
        </div>
        <button onClick={load} className="btn-secondary flex items-center gap-1.5 py-1.5 text-xs">
          <RefreshCw className="w-3.5 h-3.5" />
          Refresh
        </button>
      </div>

      {/* Filters + status tabs */}
      <div className="flex flex-wrap items-center gap-2 mb-4">
        {[
          { placeholder: 'Filter by agent…', value: agentFilter, set: (v: string) => { setAgentFilter(v); setPage(1); } },
          { placeholder: 'Filter by user…',  value: userFilter,  set: (v: string) => { setUserFilter(v);  setPage(1); } },
        ].map(({ placeholder, value, set }) => (
          <div key={placeholder} className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-600 pointer-events-none" />
            <input
              type="text"
              placeholder={placeholder}
              value={value}
              onChange={e => set(e.target.value)}
              className="w-52 pl-8 pr-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                         text-[12px] text-slate-300 placeholder-slate-600
                         focus:ring-1 focus:ring-primary-600 focus:border-primary-600 outline-none
                         transition-all duration-150"
            />
          </div>
        ))}

        {/* Status tabs */}
        <div className="flex items-center ml-auto bg-surface-800 border border-surface-500 rounded-lg p-0.5">
          {(['all', 'success', 'error', 'timeout'] as StatusFilter[]).map(s => (
            <button
              key={s}
              onClick={() => setStatus(s)}
              className={`px-3 py-1 rounded-md text-[11px] font-medium capitalize transition-colors ${
                status === s
                  ? s === 'all'     ? 'bg-primary-600 text-white'
                  : s === 'success' ? 'bg-emerald-700/60 text-emerald-300'
                  : s === 'error'   ? 'bg-red-700/60 text-red-300'
                  :                   'bg-amber-700/60 text-amber-300'
                  : 'text-slate-500 hover:text-slate-300'
              }`}
            >
              {s === 'all' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
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
        <>
          <div className="table-container">
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  {['Trace ID','Agent','User','Model','Spans','Cost','Duration','Status','Started'].map(h => (
                    <th key={h} className={TH}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {visible.length === 0 && (
                  <tr>
                    <td colSpan={9} className="text-center py-14 text-slate-600 text-sm">
                      No traces found
                    </td>
                  </tr>
                )}
                {visible.map(t => (
                  <tr key={t.trace_id} className="table-row">
                    <td className={TD}>
                      <Link
                        to={`/traces/${t.trace_id}`}
                        className="text-primary-400 hover:text-primary-300 hover:underline font-mono text-xs"
                      >
                        {t.trace_id.slice(0, 8)}…
                      </Link>
                    </td>
                    <td className={`${TD} text-slate-300`}>{t.agent_name ?? <span className="text-slate-600">—</span>}</td>
                    <td className={`${TD} text-slate-400`}>{t.user_id    ?? <span className="text-slate-600">—</span>}</td>
                    <td className={TD}>
                      {t.model
                        ? <span className={`inline-block px-2 py-0.5 rounded text-[11px] font-medium ${modelBadgeClass(t.model)}`}>{t.model}</span>
                        : <span className="text-slate-600">—</span>}
                    </td>
                    <td className={`${TD} text-slate-400 tabular-nums`}>{t.span_count}</td>
                    <td className={`${TD} text-primary-400 font-semibold tabular-nums`}>${t.total_cost_usd.toFixed(4)}</td>
                    <td className={`${TD} text-slate-400 tabular-nums`}>{t.duration_ms > 0 ? `${t.duration_ms}ms` : '—'}</td>
                    <td className={TD}>
                      <span className={`badge ${
                        t.status === 'success' ? 'badge-success' :
                        t.status === 'error'   ? 'badge-error'   : 'badge-warning'
                      }`}>{t.status}</span>
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
          <div className="flex items-center justify-between mt-3 text-xs text-slate-500">
            <span>{total} total</span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
                className="px-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                           hover:bg-surface-700 disabled:opacity-30 disabled:cursor-not-allowed
                           transition-colors text-slate-400"
              >
                ← Prev
              </button>
              <span className="px-2 font-medium text-slate-400">Page {page} / {totalPages}</span>
              <button
                disabled={traces.length < 20}
                onClick={() => setPage(p => p + 1)}
                className="px-3 py-1.5 bg-surface-800 border border-surface-500 rounded-lg
                           hover:bg-surface-700 disabled:opacity-30 disabled:cursor-not-allowed
                           transition-colors text-slate-400"
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
