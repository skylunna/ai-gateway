import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16 text-gray-400 text-sm">
        Loading…
      </div>
    );
  }
  if (error) {
    return (
      <div className="text-red-600 bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-sm">
        Error: {error}
      </div>
    );
  }
  if (!data) return null;

  const { summary, spans, timeline } = data;

  return (
    <div className="animate-fade-in">
      {/* Back link */}
      <Link
        to="/traces"
        className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-800
                   mb-4 transition-colors duration-150"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to traces
      </Link>

      {/* Trace ID */}
      <h1 className="font-mono text-sm font-medium text-gray-600 mb-5 break-all">{traceId}</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <div className="stat-card">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Spans</p>
          <p className="text-2xl font-bold text-gray-900">{summary.total_spans}</p>
        </div>
        <div className="stat-card">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Total Cost</p>
          <p className="text-2xl font-bold text-primary-600">${summary.total_cost_usd.toFixed(6)}</p>
        </div>
        <div className="stat-card">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Duration</p>
          <p className="text-2xl font-bold text-gray-900">
            {summary.duration_ms > 0 ? `${summary.duration_ms}ms` : '—'}
          </p>
        </div>
        <div className="stat-card">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Status</p>
          <div className="mt-1.5">
            <span className={`badge ${summary.status === 'success' ? 'badge-success' : 'badge-error'}`}>
              {summary.status}
            </span>
          </div>
        </div>
      </div>

      <SpanTree spans={spans} timeline={timeline} />
    </div>
  );
}
