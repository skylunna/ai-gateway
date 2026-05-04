import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Search, RefreshCw } from 'lucide-react';
import { fetchTraces } from '../api/client';
import type { TraceItem } from '../types/trace';
import { format } from 'date-fns';

export function TraceList() {
  const [traces, setTraces] = useState<TraceItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [agentFilter, setAgentFilter] = useState('');
  const [userFilter, setUserFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(() => {
    setLoading(true);
    setError('');
    fetchTraces({
      page,
      page_size: 20,
      agent_name: agentFilter || undefined,
      user_id: userFilter || undefined,
    })
      .then(data => {
        setTraces(data.traces ?? []);
        setTotal(data.total_count ?? 0);
      })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [page, agentFilter, userFilter]);

  useEffect(() => { load(); }, [load]);

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-4 mb-6">
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">Traces</h1>
          <p className="text-sm text-gray-500 mt-0.5">{total} traces total</p>
        </div>
        <button onClick={load} className="btn-secondary flex items-center gap-2 self-start sm:self-auto">
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-4 flex-wrap">
        <div className="relative flex-1 min-w-[180px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
          <input
            type="text"
            placeholder="Filter by agent…"
            value={agentFilter}
            onChange={e => { setAgentFilter(e.target.value); setPage(1); }}
            className="w-full pl-9 pr-4 py-2 border border-gray-300 rounded-lg text-sm
                       focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none
                       transition-all duration-150 bg-white"
          />
        </div>
        <div className="relative flex-1 min-w-[180px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
          <input
            type="text"
            placeholder="Filter by user…"
            value={userFilter}
            onChange={e => { setUserFilter(e.target.value); setPage(1); }}
            className="w-full pl-9 pr-4 py-2 border border-gray-300 rounded-lg text-sm
                       focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none
                       transition-all duration-150 bg-white"
          />
        </div>
      </div>

      {error && (
        <div className="text-red-600 bg-red-50 border border-red-200 rounded-lg px-4 py-3 mb-4 text-sm">
          Error: {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-16 text-gray-400 text-sm">
          Loading…
        </div>
      ) : (
        <>
          <div className="table-container">
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  {['Trace ID','Agent','User','Model','Spans','Cost (USD)','Duration','Status','Started'].map(h => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500
                                          uppercase tracking-wider bg-gray-50 border-b border-gray-200">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {traces.length === 0 && (
                  <tr>
                    <td colSpan={9} className="text-center py-16 text-gray-400 text-sm">
                      No traces found
                    </td>
                  </tr>
                )}
                {traces.map(t => (
                  <tr key={t.trace_id} className="table-row">
                    <td className="px-4 py-3">
                      <Link
                        to={`/traces/${t.trace_id}`}
                        className="text-primary-600 hover:text-primary-800 hover:underline font-mono text-xs"
                      >
                        {t.trace_id.slice(0, 8)}…
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {t.agent_name ?? <span className="text-gray-400">—</span>}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {t.user_id ?? <span className="text-gray-400">—</span>}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {t.model ?? <span className="text-gray-400">—</span>}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">{t.span_count}</td>
                    <td className="px-4 py-3 text-sm text-primary-600 font-medium">
                      ${t.total_cost_usd.toFixed(6)}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {t.duration_ms > 0 ? `${t.duration_ms}ms` : '—'}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`badge ${t.status === 'success' ? 'badge-success' : 'badge-error'}`}>
                        {t.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
                      {format(new Date(t.start_time), 'MM-dd HH:mm:ss')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between mt-4 text-sm text-gray-500">
            <span>{total} total</span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
                className="px-3 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50
                           disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-sm"
              >
                ← Prev
              </button>
              <span className="px-3 font-medium text-gray-700">Page {page}</span>
              <button
                disabled={traces.length < 20}
                onClick={() => setPage(p => p + 1)}
                className="px-3 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50
                           disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-sm"
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
