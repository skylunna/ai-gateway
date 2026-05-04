import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { fetchTrace } from '../api/client';
import type { TraceDetailResponse } from '../types/trace';
import { SpanTree } from '../components/SpanTree';

export function TraceDetail() {
  const { traceId } = useParams<{ traceId: string }>();
  const [data,    setData]    = useState<TraceDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error,   setError]   = useState('');

  useEffect(() => {
    if (!traceId) return;
    fetchTrace(traceId)
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [traceId]);

  if (loading) return (
    <div className="flex items-center justify-center py-20 text-slate-600 text-sm">Loading…</div>
  );
  if (error) return (
    <div className="text-red-400 bg-red-900/20 border border-red-800/40 rounded-xl px-4 py-3 text-sm">{error}</div>
  );
  if (!data) return null;

  const { summary, spans, timeline } = data;

  return (
    <div className="animate-fade-in">
      {/* Page header */}
      <div className="flex items-center gap-3 mb-5">
        <Link
          to="/traces"
          className="inline-flex items-center justify-center w-7 h-7 rounded-lg
                     bg-surface-800 border border-surface-500 text-slate-400
                     hover:text-slate-200 hover:bg-surface-700 transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
        </Link>
        <div>
          <h1 className="text-xl font-bold text-slate-100">Trace Detail</h1>
          <p className="text-xs text-slate-500 mt-0.5">Full span timeline and cost breakdown</p>
        </div>
      </div>

      {/* Trace ID */}
      <p className="text-xs text-slate-500 mb-5 font-mono">
        Trace ID:{' '}
        <span className="text-primary-400 font-semibold">{traceId}</span>
      </p>

      {/* Summary cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5">
        <SummaryCard label="SPANS"      value={String(summary.total_spans)} />
        <SummaryCard label="TOTAL COST" value={`$${summary.total_cost_usd.toFixed(4)}`} valueClass="text-primary-400" />
        <SummaryCard label="DURATION"   value={summary.duration_ms > 0 ? `${summary.duration_ms}ms` : '—'} />
        <SummaryCard
          label="STATUS"
          value=""
          extra={
            <span className={`badge ${summary.status === 'success' ? 'badge-success' : 'badge-error'}`}>
              {summary.status}
            </span>
          }
        />
      </div>

      <SpanTree spans={spans} timeline={timeline} />
    </div>
  );
}

interface SummaryCardProps {
  label: string;
  value: string;
  valueClass?: string;
  extra?: React.ReactNode;
}

function SummaryCard({ label, value, valueClass = 'text-white', extra }: SummaryCardProps) {
  return (
    <div className="bg-surface-800 border border-surface-500 rounded-xl p-5">
      <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-2">{label}</p>
      {extra ?? (
        <p className={`text-2xl font-bold tabular-nums ${valueClass}`}>{value}</p>
      )}
    </div>
  );
}
