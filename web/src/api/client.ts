import type {
  TraceListResponse,
  TraceDetailResponse,
  DashboardSummary,
  CostBreakdown,
  LiveMetrics,
  Policy,
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

export function fetchLiveMetrics(): Promise<LiveMetrics> {
  return get<LiveMetrics>('/metrics/live');
}

export function fetchPolicies(): Promise<Policy[]> {
  return get<Policy[]>('/policies');
}

export interface CreatePolicyRequest {
  name: string;
  expression: string;
  action: 'block' | 'alert' | 'downgrade';
  priority: number;
  description: string;
  enabled: boolean;
  tenant_id?: string;
}

export async function createPolicy(req: CreatePolicyRequest): Promise<Policy> {
  const res = await fetch(`${BASE}/policies`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? res.statusText);
  }
  return res.json() as Promise<Policy>;
}

export async function deletePolicy(id: string): Promise<void> {
  const res = await fetch(`${BASE}/policies/${id}`, { method: 'DELETE' });
  if (!res.ok && res.status !== 204) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? res.statusText);
  }
}
