import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
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
    fetchTraces({ page, page_size: 20, agent_name: agentFilter || undefined, user_id: userFilter || undefined })
      .then(data => {
        setTraces(data.traces ?? []);
        setTotal(data.total_count ?? 0);
      })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [page, agentFilter, userFilter]);

  useEffect(() => { load(); }, [load]);

  return (
    <>
      <h1>Traces</h1>

      <div className="filters">
        <input
          placeholder="Filter by agent"
          value={agentFilter}
          onChange={e => { setAgentFilter(e.target.value); setPage(1); }}
        />
        <input
          placeholder="Filter by user"
          value={userFilter}
          onChange={e => { setUserFilter(e.target.value); setPage(1); }}
        />
        <button onClick={load}>Refresh</button>
      </div>

      {error && <div className="empty">Error: {error}</div>}
      {loading && <div className="loading">Loading…</div>}
      {!loading && !error && (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Trace ID</th>
                  <th>Agent</th>
                  <th>User</th>
                  <th>Model</th>
                  <th>Spans</th>
                  <th>Cost (USD)</th>
                  <th>Duration</th>
                  <th>Status</th>
                  <th>Started</th>
                </tr>
              </thead>
              <tbody>
                {traces.length === 0 && (
                  <tr><td colSpan={9} className="empty">No traces found</td></tr>
                )}
                {traces.map(t => (
                  <tr key={t.trace_id}>
                    <td>
                      <Link to={`/traces/${t.trace_id}`} className="trace-link">
                        {t.trace_id.slice(0, 8)}…
                      </Link>
                    </td>
                    <td>{t.agent_name ?? <span style={{ color: '#aaa' }}>—</span>}</td>
                    <td>{t.user_id ?? <span style={{ color: '#aaa' }}>—</span>}</td>
                    <td>{t.model ?? <span style={{ color: '#aaa' }}>—</span>}</td>
                    <td>{t.span_count}</td>
                    <td>${t.total_cost_usd.toFixed(6)}</td>
                    <td>{t.duration_ms > 0 ? `${t.duration_ms}ms` : '—'}</td>
                    <td>
                      <span className={`badge badge-${t.status}`}>{t.status}</span>
                    </td>
                    <td>{format(new Date(t.start_time), 'MM-dd HH:mm:ss')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="pagination">
            <span>{total} total</span>
            <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>← Prev</button>
            <span>Page {page}</span>
            <button disabled={traces.length < 20} onClick={() => setPage(p => p + 1)}>Next →</button>
          </div>
        </>
      )}
    </>
  );
}
