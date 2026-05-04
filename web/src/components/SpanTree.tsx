import type { SpanNode as SpanNodeData, Timeline } from '../types/trace';
import { SpanNode } from './SpanNode';

interface Props {
  spans: SpanNodeData[];
  timeline: Timeline;
}

export function SpanTree({ spans, timeline }: Props) {
  if (!spans || spans.length === 0) {
    return (
      <div className="flex items-center justify-center py-12 text-gray-400 text-sm
                      bg-white rounded-xl border border-gray-200 shadow-card">
        No spans found
      </div>
    );
  }

  const dur = timeline.duration_ms;

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-card overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200">
        <span className="text-sm font-semibold text-gray-800">Span Timeline</span>
        <span className="text-xs text-gray-400">total {dur}ms</span>
      </div>

      {/* Time ruler */}
      <div className="flex justify-between px-4 py-1.5 text-[10px] text-gray-300 border-b border-gray-100"
           style={{ paddingLeft: '2.5rem' }}>
        <span>0</span>
        <span>{Math.round(dur / 4)}ms</span>
        <span>{Math.round(dur / 2)}ms</span>
        <span>{Math.round((dur * 3) / 4)}ms</span>
        <span>{dur}ms</span>
      </div>

      {/* Span rows */}
      <div>
        {spans.map(span => (
          <SpanNode
            key={span.span_id}
            span={span}
            depth={0}
            totalDuration={dur}
          />
        ))}
      </div>
    </div>
  );
}
