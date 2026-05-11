import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import {
  AreaChart, Area, BarChart, Bar,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts';
import { Sparkles, Layers, DollarSign, Clock, AlertTriangle, RefreshCw, Database, TrendingUp, XCircle, Trash2 } from 'lucide-react';
import { fetchDashboardSummary, fetchCostBreakdown, fetchLiveMetrics } from '../api/client';
import type { DashboardSummary, CostBreakdown, CostItem, LiveMetrics } from '../types/trace';

const TIME_RANGES = ['1h', '6h', '24h', '7d', '30d'] as const;
type TimeRange = typeof TIME_RANGES[number];

/* ── Static mock chart data ──────────────────────────────── */

const LATENCY_DATA = [
  { h:'00:00',p50:340,p99:820 }, { h:'01:00',p50:310,p99:770 },
  { h:'02:00',p50:290,p99:730 }, { h:'03:00',p50:275,p99:700 },
  { h:'04:00',p50:270,p99:690 }, { h:'05:00',p50:285,p99:715 },
  { h:'06:00',p50:320,p99:790 }, { h:'07:00',p50:380,p99:910 },
  { h:'08:00',p50:440,p99:1050}, { h:'09:00',p50:490,p99:1160},
  { h:'10:00',p50:510,p99:1210}, { h:'11:00',p50:495,p99:1180},
  { h:'12:00',p50:460,p99:1100}, { h:'13:00',p50:465,p99:1110},
  { h:'14:00',p50:445,p99:1070}, { h:'15:00',p50:425,p99:1010},
  { h:'16:00',p50:435,p99:1030}, { h:'17:00',p50:415,p99:990 },
  { h:'18:00',p50:420,p99:1000}, { h:'19:00',p50:395,p99:950 },
  { h:'20:00',p50:375,p99:905 }, { h:'21:00',p50:355,p99:860 },
  { h:'22:00',p50:335,p99:820 }, { h:'23:00',p50:315,p99:780 },
];

const REQUESTS_DATA = [
  { h:'00:00',n:28 }, { h:'01:00',n:22 }, { h:'02:00',n:18 }, { h:'03:00',n:15 },
  { h:'04:00',n:14 }, { h:'05:00',n:20 }, { h:'06:00',n:35 }, { h:'07:00',n:52 },
  { h:'08:00',n:71 }, { h:'09:00',n:84 }, { h:'10:00',n:89 }, { h:'11:00',n:82 },
  { h:'12:00',n:75 }, { h:'13:00',n:78 }, { h:'14:00',n:80 }, { h:'15:00',n:73 },
  { h:'16:00',n:68 }, { h:'17:00',n:62 }, { h:'18:00',n:55 }, { h:'19:00',n:48 },
  { h:'20:00',n:42 }, { h:'21:00',n:38 }, { h:'22:00',n:33 }, { h:'23:00',n:29 },
];

const TOKEN_DATA = [
  { day:'Apr 28', prompt:85000,  completion:32000 },
  { day:'Apr 29', prompt:92000,  completion:35000 },
  { day:'Apr 30', prompt:78000,  completion:29000 },
  { day:'May 1',  prompt:95000,  completion:38000 },
  { day:'May 2',  prompt:105000, completion:42000 },
  { day:'May 3',  prompt:98000,  completion:39000 },
  { day:'May 4',  prompt:88000,  completion:34000 },
];

/* ── Shared chart style constants ─────────────────────────── */

const GRID  = { strokeDasharray:'3 3', stroke:'#1e2b3d', vertical: false } as const;
const XAXIS = { tick:{ fill:'#475569', fontSize:10 }, tickLine:false, axisLine:false } as const;
const YAXIS = { tick:{ fill:'#475569', fontSize:10 }, tickLine:false, axisLine:false } as const;
const TT    = {
  contentStyle: {
    backgroundColor:'#1a2440',
    border:'1px solid #2a3c5a',
    borderRadius:'8px',
    fontSize:11,
    color:'#cbd5e1',
  },
  cursor:{ stroke:'#2a3c5a', strokeWidth:1 },
} as const;

/* ── Legend dot ───────────────────────────────────────────── */
function Dot({ color, label }: { color: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5 text-[11px] text-slate-500">
      <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ background: color }} />
      {label}
    </span>
  );
}

/* ── Main page ────────────────────────────────────────────── */

export function Dashboard() {
  const [summary,   setSummary]   = useState<DashboardSummary | null>(null);
  const [breakdown, setBreakdown] = useState<CostBreakdown | null>(null);
  const [live,      setLive]      = useState<LiveMetrics | null>(null);
  const [timeRange, setTimeRange] = useState<TimeRange>('24h');
  const [tick,      setTick]      = useState(0);

  const refresh = () => setTick(t => t + 1);

  useEffect(() => {
    setSummary(null);
    setBreakdown(null);
    Promise.allSettled([fetchDashboardSummary(), fetchCostBreakdown()])
      .then(([s, b]) => {
        if (s.status === 'fulfilled') setSummary(s.value);
        if (b.status === 'fulfilled') setBreakdown(b.value);
      });
  }, [tick]);

  // poll live metrics every 5 s
  useEffect(() => {
    let alive = true;
    const poll = () => fetchLiveMetrics().then(m => { if (alive) setLive(m); }).catch(() => {});
    poll();
    const id = setInterval(poll, 5000);
    return () => { alive = false; clearInterval(id); };
  }, []);

  return (
    <div className="animate-fade-in space-y-4">
      {/* ── Page header ── */}
      <div className="flex items-start justify-between mb-2">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Dashboard</h1>
          <p className="text-xs text-slate-500 mt-0.5">LLM gateway overview · real-time metrics</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Time range tabs */}
          <div className="flex items-center bg-surface-800 border border-surface-500 rounded-lg p-0.5">
            {TIME_RANGES.map(r => (
              <button
                key={r}
                onClick={() => setTimeRange(r)}
                className={`px-2.5 py-1 rounded-md text-[11px] font-medium transition-colors ${
                  timeRange === r
                    ? 'bg-primary-600 text-white'
                    : 'text-slate-500 hover:text-slate-300'
                }`}
              >
                {r}
              </button>
            ))}
          </div>
          <button onClick={refresh} className="btn-secondary flex items-center gap-1.5 py-1.5 text-xs">
            <RefreshCw className="w-3.5 h-3.5" />
            Refresh
          </button>
        </div>
      </div>

      {/* ── Stat cards ──── */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        <StatCard
          icon={<Sparkles className="w-4 h-4" />}
          iconBg="bg-blue-500/15" iconColor="text-blue-400"
          label="TRACES"
          value={summary?.total_traces.toLocaleString() ?? '—'}
          trend="+12% vs prev" trendColor="text-emerald-400"
        />
        <StatCard
          icon={<Layers className="w-4 h-4" />}
          iconBg="bg-violet-500/15" iconColor="text-violet-400"
          label="SPANS"
          value={summary?.total_spans.toLocaleString() ?? '—'}
          trend="+8% vs prev" trendColor="text-emerald-400"
        />
        <StatCard
          icon={<DollarSign className="w-4 h-4" />}
          iconBg="bg-blue-500/15" iconColor="text-blue-400"
          label="TOTAL COST"
          value={summary ? `$${summary.total_cost_usd.toFixed(2)}` : '—'}
          trend="+3% vs prev" trendColor="text-amber-400"
        />
        <StatCard
          icon={<Clock className="w-4 h-4" />}
          iconBg="bg-emerald-500/15" iconColor="text-emerald-400"
          label="AVG LATENCY"
          value={summary ? `${Math.round(summary.avg_latency_ms)}ms` : '—'}
          trend="−5% vs prev" trendColor="text-emerald-400"
        />
        <StatCard
          icon={<AlertTriangle className="w-4 h-4" />}
          iconBg="bg-red-500/15" iconColor="text-red-400"
          label="ERROR RATE"
          value={summary ? `${(summary.error_rate * 100).toFixed(1)}%` : '—'}
          trend="+0.2% vs prev" trendColor="text-red-400"
        />
      </div>

      {/* ── Row 2: Latency (2/3) + Requests (1/3) ── */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Latency P50/P99 area chart */}
        <div className="chart-card lg:col-span-2">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="chart-title">Latency (P50 / P99)</p>
              <p className="chart-sub">Hourly · ms · last 24h</p>
            </div>
            <div className="flex items-center gap-4">
              <Dot color="#3b82f6" label="P50" />
              <Dot color="#a855f7" label="P99" />
            </div>
          </div>
          <ResponsiveContainer width="100%" height={180}>
            <AreaChart data={LATENCY_DATA} margin={{ top:4, right:4, bottom:0, left:0 }}>
              <defs>
                <linearGradient id="gP99" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%"   stopColor="#a855f7" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#a855f7" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="gP50" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%"   stopColor="#3b82f6" stopOpacity={0.35} />
                  <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid {...GRID} />
              <XAxis dataKey="h" {...XAXIS} ticks={['00:00','06:00','12:00','18:00']} />
              <YAxis {...YAXIS} width={45} />
              <Tooltip {...TT} formatter={(v: unknown) => [`${v}ms`]} />
              <Area type="monotone" dataKey="p99" stroke="#a855f7" strokeWidth={1.5} fill="url(#gP99)" dot={false} activeDot={{ r:3, fill:'#a855f7' }} />
              <Area type="monotone" dataKey="p50" stroke="#3b82f6" strokeWidth={1.5} fill="url(#gP50)" dot={false} activeDot={{ r:3, fill:'#3b82f6' }} />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Requests / hour bar chart */}
        <div className="chart-card">
          <div className="mb-4">
            <p className="chart-title">Requests / hour</p>
            <p className="chart-sub">24h window</p>
          </div>
          <ResponsiveContainer width="100%" height={180}>
            <BarChart data={REQUESTS_DATA} margin={{ top:4, right:4, bottom:0, left:0 }}>
              <CartesianGrid {...GRID} />
              <XAxis dataKey="h" {...XAXIS} ticks={['00:00','08:00','16:00']} />
              <YAxis {...YAXIS} width={28} />
              <Tooltip {...TT} formatter={(v: unknown) => [`${v} req`]} />
              <Bar dataKey="n" fill="#3b82f6" radius={[2,2,0,0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* ── Row 3: Tokens (1/2) + Cost by agent (1/2) ── */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Token consumption stacked bar */}
        <div className="chart-card">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="chart-title">Token Consumption</p>
              <p className="chart-sub">7-day · prompt vs completion</p>
            </div>
            <div className="flex items-center gap-4">
              <Dot color="#3b82f6" label="Prompt" />
              <Dot color="#22d3ee" label="Completion" />
            </div>
          </div>
          <ResponsiveContainer width="100%" height={180}>
            <BarChart data={TOKEN_DATA} margin={{ top:4, right:4, bottom:0, left:0 }}>
              <CartesianGrid {...GRID} />
              <XAxis dataKey="day" {...XAXIS} />
              <YAxis {...YAXIS} width={42} tickFormatter={v => v >= 1000 ? `${(v as number)/1000}k` : String(v)} />
              <Tooltip {...TT} formatter={(v: unknown, name: unknown) => [`${((v as number)/1000).toFixed(0)}k`, name === 'prompt' ? 'Prompt' : 'Completion']} />
              <Bar dataKey="prompt"     fill="#3b82f6" stackId="t" radius={[0,0,0,0]} />
              <Bar dataKey="completion" fill="#22d3ee" stackId="t" radius={[2,2,0,0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Cost by agent horizontal bars */}
        <CostByAgent items={breakdown?.by_agent ?? []} />
      </div>

      {/* ── Row 4: Cache & Live Metrics ── */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <h2 className="text-sm font-semibold text-slate-300">Cache &amp; Live Metrics</h2>
          <span className="flex items-center gap-1 text-[10px] text-emerald-400 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
            live · 5s
          </span>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <LiveStat
            icon={<TrendingUp className="w-4 h-4" />}
            iconBg="bg-emerald-500/15" iconColor="text-emerald-400"
            label="HIT RATE"
            value={live ? `${(live.cache.hit_rate * 100).toFixed(1)}%` : '—'}
          />
          <LiveStat
            icon={<Database className="w-4 h-4" />}
            iconBg="bg-blue-500/15" iconColor="text-blue-400"
            label="CACHE SIZE"
            value={live ? live.cache.size.toLocaleString() : '—'}
          />
          <LiveStat
            icon={<Sparkles className="w-4 h-4" />}
            iconBg="bg-violet-500/15" iconColor="text-violet-400"
            label="HITS"
            value={live ? live.cache.hits.toLocaleString() : '—'}
          />
          <LiveStat
            icon={<XCircle className="w-4 h-4" />}
            iconBg="bg-amber-500/15" iconColor="text-amber-400"
            label="MISSES"
            value={live ? live.cache.misses.toLocaleString() : '—'}
          />
          <LiveStat
            icon={<Trash2 className="w-4 h-4" />}
            iconBg="bg-red-500/15" iconColor="text-red-400"
            label="LRU EVICTIONS"
            value={live ? (live.cache.evictions['capacity'] ?? 0).toLocaleString() : '—'}
          />
          <LiveStat
            icon={<Clock className="w-4 h-4" />}
            iconBg="bg-slate-500/15" iconColor="text-slate-400"
            label="TTL EXPIRED"
            value={live ? (live.cache.evictions['ttl'] ?? 0).toLocaleString() : '—'}
          />
        </div>

        {/* Request & token totals */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
          <RequestsByStatus data={live?.requests_by_status ?? {}} />
          <TokensByType data={live?.tokens_by_type ?? {}} />
        </div>
      </div>
    </div>
  );
}

/* ── StatCard ─────────────────────────────────────────────── */

interface StatCardProps {
  icon: ReactNode;
  iconBg: string;
  iconColor: string;
  label: string;
  value: string;
  trend: string;
  trendColor: string;
}

function StatCard({ icon, iconBg, iconColor, label, value, trend, trendColor }: StatCardProps) {
  return (
    <div className="stat-card">
      <div className={`inline-flex items-center justify-center w-9 h-9 rounded-lg mb-3 ${iconBg} ${iconColor}`}>
        {icon}
      </div>
      <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">{label}</p>
      <p className="text-[26px] font-bold text-white leading-none mb-1.5 tabular-nums">{value}</p>
      <p className={`text-[11px] font-medium ${trendColor}`}>{trend}</p>
    </div>
  );
}

/* ── LiveStat (no trend line, for live data) ─────────────── */

interface LiveStatProps {
  icon: ReactNode;
  iconBg: string;
  iconColor: string;
  label: string;
  value: string;
}

function LiveStat({ icon, iconBg, iconColor, label, value }: LiveStatProps) {
  return (
    <div className="stat-card">
      <div className={`inline-flex items-center justify-center w-9 h-9 rounded-lg mb-3 ${iconBg} ${iconColor}`}>
        {icon}
      </div>
      <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">{label}</p>
      <p className="text-[22px] font-bold text-white leading-none tabular-nums">{value}</p>
    </div>
  );
}

/* ── RequestsByStatus ─────────────────────────────────────── */

function RequestsByStatus({ data }: { data: Record<string, number> }) {
  const entries = Object.entries(data).sort((a, b) => b[1] - a[1]);
  const total = entries.reduce((s, [, v]) => s + v, 0);

  const color = (status: string) => {
    if (status === '200-cache') return 'bg-emerald-500';
    if (status.startsWith('2'))  return 'bg-blue-500';
    if (status.startsWith('4'))  return 'bg-amber-500';
    return 'bg-red-500';
  };

  return (
    <div className="chart-card">
      <p className="chart-title mb-1">Requests by Status</p>
      <p className="chart-sub mb-4">cumulative · all time</p>
      {total === 0 ? (
        <p className="text-slate-600 text-sm">No data yet</p>
      ) : (
        <div className="space-y-2.5">
          {entries.map(([status, count]) => (
            <div key={status}>
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-mono text-slate-400">{status}</span>
                <span className="text-xs font-semibold text-slate-300 tabular-nums">
                  {count.toLocaleString()}
                </span>
              </div>
              <div className="h-[3px] bg-surface-600 rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-500 ${color(status)}`}
                  style={{ width: `${(count / total) * 100}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ── TokensByType ─────────────────────────────────────────── */

function TokensByType({ data }: { data: Record<string, number> }) {
  const order = ['prompt', 'completion', 'total'];
  const entries = [...Object.entries(data)].sort(
    (a, b) => order.indexOf(a[0]) - order.indexOf(b[0]),
  );

  const colorClass: Record<string, string> = {
    prompt:     'text-blue-400',
    completion: 'text-cyan-400',
    total:      'text-violet-400',
  };

  return (
    <div className="chart-card">
      <p className="chart-title mb-1">Tokens by Type</p>
      <p className="chart-sub mb-4">cumulative · all time</p>
      {entries.length === 0 ? (
        <p className="text-slate-600 text-sm">No data yet</p>
      ) : (
        <div className="grid grid-cols-3 gap-3">
          {entries.map(([type, count]) => (
            <div key={type} className="bg-surface-800 rounded-lg p-3 text-center">
              <p className={`text-[11px] font-semibold uppercase tracking-wider mb-1 ${colorClass[type] ?? 'text-slate-400'}`}>
                {type}
              </p>
              <p className="text-lg font-bold text-white tabular-nums">
                {count >= 1_000_000
                  ? `${(count / 1_000_000).toFixed(1)}M`
                  : count >= 1_000
                  ? `${(count / 1_000).toFixed(1)}k`
                  : count.toString()}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ── CostByAgent ──────────────────────────────────────────── */

function CostByAgent({ items }: { items: CostItem[] }) {
  const maxCost = Math.max(...items.map(i => i.cost), 0.001);

  return (
    <div className="chart-card flex flex-col">
      <div className="mb-5">
        <p className="chart-title">Cost by Agent</p>
        <p className="chart-sub">USD · last 24h</p>
      </div>

      {items.length === 0 ? (
        <div className="flex-1 flex items-center justify-center text-slate-600 text-sm">No data</div>
      ) : (
        <div className="space-y-3.5">
          {items.map(item => (
            <div key={item.name}>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-sm text-slate-400 font-mono">{item.name}</span>
                <span className="text-sm font-semibold text-primary-400 tabular-nums">
                  ${item.cost.toFixed(4)}
                </span>
              </div>
              <div className="h-[3px] bg-surface-600 rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary-500 rounded-full transition-all duration-500"
                  style={{ width: `${(item.cost / maxCost) * 100}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
