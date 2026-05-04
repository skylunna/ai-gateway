import type { SpanNode as SpanNodeData, Timeline } from '../types/trace';
import { SpanNode } from './SpanNode';

interface Props {
  spans: SpanNodeData[];
  timeline: Timeline;
}

export function SpanTree({ spans, timeline }: Props) {
  if (!spans || spans.length === 0) {
    return (
      <div className="flex items-center justify-center py-12 text-slate-600 text-sm
                      bg-surface-800 border border-surface-500 rounded-xl">
        No spans found
      </div>
    );
  }

  const dur = timeline.duration_ms;

  return (
    <div className="bg-surface-800 border border-surface-500 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-surface-900 border-b border-surface-600">
        <span className="text-sm font-semibold text-slate-200">Span Timeline</span>
        <span className="text-xs font-mono text-slate-500">total {dur}ms</span>
      </div>

      {/* Time ruler */}
      <div
        className="flex justify-between px-4 py-2 text-[10px] text-slate-600
                   border-b border-surface-600/50 bg-surface-900/50 font-mono"
        style={{ paddingLeft: '3rem' }}
      >
        {[0, 0.25, 0.5, 0.75, 1].map(frac => (
          <span key={frac}>{Math.round(dur * frac)}ms</span>
        ))}
      </div>

      {/* Span rows */}
      <div>
        {spans.map(span => (
          <SpanNode key={span.span_id} span={span} depth={0} totalDuration={dur} />
        ))}
      </div>
    </div>
  );
}
