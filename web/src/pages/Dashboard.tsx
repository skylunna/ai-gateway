import { useEffect, useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { fetchDashboardSummary, fetchCostBreakdown } from '../api/client';
import type { DashboardSummary, CostBreakdown } from '../types/trace';

export function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [breakdown, setBreakdown] = useState<CostBreakdown | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([fetchDashboardSummary(), fetchCostBreakdown()])
      .then(([s, b]) => { setSummary(s); setBreakdown(b); })
      .catch(e => setError(e.message));
  }, []);

  if (error) return <div className="empty">Error: {error}</div>;

  return (
    <>
      <h1>Dashboard <span style={{ fontSize: 13, color: '#888', fontWeight: 400 }}>last 24 h</span></h1>

      <div className="stat-grid">
        <div className="stat-card">
          <div className="stat-label">Traces</div>
          <div className="stat-value">{summary?.total_traces ?? '…'}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Spans</div>
          <div className="stat-value">{summary?.total_spans ?? '…'}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total Cost</div>
          <div className="stat-value cost">
            {summary ? `$${summary.total_cost_usd.toFixed(4)}` : '…'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Avg Latency</div>
          <div className="stat-value">
            {summary ? `${Math.round(summary.avg_latency_ms)}ms` : '…'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Error Rate</div>
          <div className={`stat-value${(summary?.error_rate ?? 0) > 0 ? ' error' : ''}`}>
            {summary ? `${(summary.error_rate * 100).toFixed(1)}%` : '…'}
          </div>
        </div>
      </div>

      <div className="charts">
        <div className="card">
          <h2>Cost by Agent</h2>
          {breakdown?.by_agent.length === 0 && <div className="empty">No data</div>}
          {breakdown && breakdown.by_agent.length > 0 && (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={breakdown.by_agent} margin={{ top: 4, right: 8, bottom: 4, left: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis tick={{ fontSize: 11 }} tickFormatter={v => `$${v.toFixed(4)}`} width={70} />
                <Tooltip formatter={(v: number) => [`$${v.toFixed(6)}`, 'Cost']} />
                <Bar dataKey="cost" fill="#2563eb" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="card">
          <h2>Cost by User</h2>
          {breakdown?.by_user.length === 0 && <div className="empty">No data</div>}
          {breakdown && breakdown.by_user.length > 0 && (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={breakdown.by_user} margin={{ top: 4, right: 8, bottom: 4, left: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis tick={{ fontSize: 11 }} tickFormatter={v => `$${v.toFixed(4)}`} width={70} />
                <Tooltip formatter={(v: number) => [`$${v.toFixed(6)}`, 'Cost']} />
                <Bar dataKey="cost" fill="#7c3aed" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </>
  );
}
