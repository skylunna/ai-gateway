import type { SpanNode as SpanNodeData, Timeline } from '../types/trace';
import { SpanNode } from './SpanNode';

interface Props {
  spans: SpanNodeData[];
  timeline: Timeline;
}

export function SpanTree({ spans, timeline }: Props) {
  if (!spans || spans.length === 0) {
    return <div className="empty">No spans found</div>;
  }

  const dur = timeline.duration_ms;

  return (
    <div className="st-wrap">
      {/* Header */}
      <div className="st-header">
        <span className="st-title">Span Timeline</span>
        <span className="st-total">total {dur}ms</span>
      </div>

      {/* Time ruler */}
      <div className="st-ruler">
        <span>0</span>
        <span>{Math.round(dur / 4)}ms</span>
        <span>{Math.round(dur / 2)}ms</span>
        <span>{Math.round((dur * 3) / 4)}ms</span>
        <span>{dur}ms</span>
      </div>

      {/* Span rows */}
      <div className="st-body">
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
