import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { fetchTrace } from '../api/client';
import type { TraceDetailResponse } from '../types/trace';
import { SpanTree } from '../components/SpanTree';

export function TraceDetail() {
  const { traceId } = useParams<{ traceId: string }>();
  const [data, setData] = useState<TraceDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!traceId) return;
    fetchTrace(traceId)
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [traceId]);

  if (loading) return <div className="loading">Loading…</div>;
  if (error)   return <div className="empty">Error: {error}</div>;
  if (!data)   return null;

  const { summary, spans, timeline } = data;

  return (
    <>
      <Link to="/traces" className="back-btn">← Back to traces</Link>
      <h1 style={{ fontFamily: 'monospace', fontSize: 15 }}>{traceId}</h1>

      <div className="stat-grid" style={{ margin: '1rem 0' }}>
        <div className="stat-card">
          <div className="stat-label">Spans</div>
          <div className="stat-value">{summary.total_spans}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total Cost</div>
          <div className="stat-value cost">${summary.total_cost_usd.toFixed(6)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Duration</div>
          <div className="stat-value">
            {summary.duration_ms > 0 ? `${summary.duration_ms}ms` : '—'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Status</div>
          <div className="stat-value">
            <span className={`badge badge-${summary.status}`}>{summary.status}</span>
          </div>
        </div>
      </div>

      <SpanTree spans={spans} timeline={timeline} />
    </>
  );
}
