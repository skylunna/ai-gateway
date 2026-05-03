export interface TraceItem {
  trace_id: string;
  agent_name?: string;
  user_id?: string;
  model?: string;
  span_count: number;
  total_cost_usd: number;
  duration_ms: number;
  start_time: string;
  status: 'success' | 'error';
}

export interface TraceListResponse {
  traces: TraceItem[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface Span {
  span_id: string;
  trace_id: string;
  parent_span_id?: string;
  session_id?: string;
  tenant_id?: string;
  user_id?: string;
  agent_name?: string;
  agent_version?: string;
  environment?: string;
  span_type: string;
  name: string;
  start_time: string;
  end_time: string;
  duration_ms: number;
  model?: string;
  prompt_tokens: number;
  completion_tokens: number;
  cost_usd: number;
  status: 'success' | 'error' | 'timeout';
  error_message?: string;
  tags?: Record<string, string>;
}

// SpanNode is the tree-structured span used in the trace detail view.
export interface SpanNode {
  span_id: string;
  parent_span_id?: string;
  name: string;
  span_type: string;
  start_time: string;
  end_time: string;
  duration_ms: number;
  relative_start_ms: number;
  relative_end_ms: number;
  model?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  cost_usd?: number;
  status: 'success' | 'error' | 'timeout';
  error_message?: string;
  children?: SpanNode[];
}

export interface Timeline {
  start_time: string;
  end_time: string;
  duration_ms: number;
}

export interface TraceSummary {
  total_spans: number;
  total_cost_usd: number;
  duration_ms: number;
  status: string;
}

export interface TraceDetailResponse {
  trace_id: string;
  spans: SpanNode[];
  summary: TraceSummary;
  timeline: Timeline;
}

export interface DashboardSummary {
  total_traces: number;
  total_spans: number;
  total_cost_usd: number;
  avg_latency_ms: number;
  error_rate: number;
}

export interface CostItem {
  name: string;
  cost: number;
  count: number;
}

export interface CostBreakdown {
  by_agent: CostItem[];
  by_user: CostItem[];
}
