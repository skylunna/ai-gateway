import type {
  TraceListResponse,
  TraceDetailResponse,
  DashboardSummary,
  CostBreakdown,
} from '../types/trace';

const BASE = '/api';

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? res.statusText);
  }
  return res.json() as Promise<T>;
}

export interface TraceQuery {
  page?: number;
  page_size?: number;
  agent_name?: string;
  user_id?: string;
}

export function fetchTraces(q: TraceQuery = {}): Promise<TraceListResponse> {
  const params = new URLSearchParams();
  if (q.page) params.set('page', String(q.page));
  if (q.page_size) params.set('page_size', String(q.page_size));
  if (q.agent_name) params.set('agent_name', q.agent_name);
  if (q.user_id) params.set('user_id', q.user_id);
  const qs = params.toString();
  return get<TraceListResponse>(`/traces${qs ? `?${qs}` : ''}`);
}

export function fetchTrace(traceId: string): Promise<TraceDetailResponse> {
  return get<TraceDetailResponse>(`/traces/${traceId}`);
}

export function fetchDashboardSummary(): Promise<DashboardSummary> {
  return get<DashboardSummary>('/dashboard/summary');
}

export function fetchCostBreakdown(): Promise<CostBreakdown> {
  return get<CostBreakdown>('/dashboard/cost');
}
