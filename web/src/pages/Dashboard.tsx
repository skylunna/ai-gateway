import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts';
import { Activity, DollarSign, Clock, AlertCircle, Zap } from 'lucide-react';
import { fetchDashboardSummary, fetchCostBreakdown } from '../api/client';
import type { DashboardSummary, CostBreakdown, CostItem } from '../types/trace';

export function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [breakdown, setBreakdown] = useState<CostBreakdown | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([fetchDashboardSummary(), fetchCostBreakdown()])
      .then(([s, b]) => { setSummary(s); setBreakdown(b); })
      .catch(e => setError(e.message));
  }, []);

  if (error) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="text-red-600 bg-red-50 border border-red-200 rounded-lg px-6 py-4 text-sm">
          Error: {error}
        </div>
      </div>
    );
  }

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-0.5">Last 24 hours</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 mb-6">
        <StatCard
          icon={<Activity className="w-5 h-5" />}
          title="Traces"
          value={summary?.total_traces.toLocaleString() ?? '…'}
          color="blue"
        />
        <StatCard
          icon={<Zap className="w-5 h-5" />}
          title="Spans"
          value={summary?.total_spans.toLocaleString() ?? '…'}
          color="indigo"
        />
        <StatCard
          icon={<DollarSign className="w-5 h-5" />}
          title="Total Cost"
          value={summary ? `$${summary.total_cost_usd.toFixed(4)}` : '…'}
          color="blue"
        />
        <StatCard
          icon={<Clock className="w-5 h-5" />}
          title="Avg Latency"
          value={summary ? `${Math.round(summary.avg_latency_ms)}ms` : '…'}
          color="purple"
        />
        <StatCard
          icon={<AlertCircle className="w-5 h-5" />}
          title="Error Rate"
          value={summary ? `${(summary.error_rate * 100).toFixed(1)}%` : '…'}
          color={(summary?.error_rate ?? 0) > 0 ? 'red' : 'green'}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <ChartCard title="Cost by Agent" data={breakdown?.by_agent ?? []} color="#2563eb" />
        <ChartCard title="Cost by User"  data={breakdown?.by_user  ?? []} color="#7c3aed" />
      </div>
    </div>
  );
}

/* ── StatCard ─────────────────────────────────────────────── */

const ICON_BG: Record<string, string> = {
  blue:   'bg-blue-100 text-blue-600',
  indigo: 'bg-indigo-100 text-indigo-600',
  purple: 'bg-purple-100 text-purple-600',
  green:  'bg-green-100 text-green-600',
  red:    'bg-red-100 text-red-600',
};

interface StatCardProps {
  icon: ReactNode;
  title: string;
  value: string;
  color: string;
}

function StatCard({ icon, title, value, color }: StatCardProps) {
  return (
    <div className="stat-card">
      <div className={`inline-flex p-2 rounded-lg mb-3 ${ICON_BG[color] ?? ICON_BG.blue}`}>
        {icon}
      </div>
      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">{title}</p>
      <p className="text-2xl font-bold text-gray-900 truncate">{value}</p>
    </div>
  );
}

/* ── ChartCard ─────────────────────────────────────────────── */

interface ChartCardProps {
  title: string;
  data: CostItem[];
  color: string;
}

function ChartCard({ title, data, color }: ChartCardProps) {
  return (
    <div className="bg-white rounded-xl p-5 shadow-card border border-gray-100">
      <h2 className="text-sm font-semibold text-gray-900 mb-4">{title}</h2>
      {data.length === 0 ? (
        <div className="flex items-center justify-center h-48 text-gray-400 text-sm">
          No data
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={220}>
          <BarChart data={data} margin={{ top: 4, right: 8, bottom: 4, left: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            <XAxis dataKey="name" tick={{ fontSize: 12, fill: '#6b7280' }} />
            <YAxis
              tick={{ fontSize: 11, fill: '#6b7280' }}
              tickFormatter={v => `$${(v as number).toFixed(4)}`}
              width={72}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'white',
                border: '1px solid #e5e7eb',
                borderRadius: '0.5rem',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
                fontSize: 12,
              }}
              formatter={(v: unknown) => [`$${(v as number).toFixed(6)}`, 'Cost']}
            />
            <Bar dataKey="cost" fill={color} radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
