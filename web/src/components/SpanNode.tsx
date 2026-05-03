import { memo, useState } from 'react';
import type { SpanNode as SpanNodeData } from '../types/trace';

const TYPE_CLASS: Record<string, string> = {
  llm:       'sn-type-llm',
  tool:      'sn-type-tool',
  retrieval: 'sn-type-retrieval',
  custom:    'sn-type-custom',
};

const TYPE_COLOR: Record<string, string> = {
  llm:       '#3b82f6',
  tool:      '#10b981',
  retrieval: '#8b5cf6',
  custom:    '#9ca3af',
};

interface Props {
  span: SpanNodeData;
  depth: number;
  totalDuration: number;
}

export const SpanNode = memo(function SpanNode({ span, depth, totalDuration }: Props) {
  const [expanded, setExpanded] = useState(true);
  const hasChildren = !!span.children && span.children.length > 0;

  const safe = totalDuration > 0 ? totalDuration : 1;
  const leftPct  = (span.relative_start_ms / safe) * 100;
  const widthPct = Math.max((span.duration_ms / safe) * 100, 0.5);

  const typeClass = TYPE_CLASS[span.span_type] ?? 'sn-type-custom';
  const barColor  = TYPE_COLOR[span.span_type] ?? '#9ca3af';
  const indent    = 1 + depth * 1.25; // rem

  return (
    <div className="sn-wrap">
      {/* ── Info row ── */}
      <div className="sn-row" style={{ paddingLeft: `${indent}rem` }}>
        <button
          className="sn-toggle"
          onClick={() => hasChildren && setExpanded(e => !e)}
          aria-label={expanded ? 'collapse' : 'expand'}
          tabIndex={hasChildren ? 0 : -1}
        >
          {hasChildren ? (expanded ? '▾' : '▸') : ''}
        </button>

        <span className={`sn-badge ${typeClass}`}>{span.span_type}</span>
        <span className="sn-name" title={span.name}>{span.name}</span>
        {span.model && <span className="sn-model">{span.model}</span>}
        <span className="sn-dur">{span.duration_ms}ms</span>
        {span.cost_usd != null && span.cost_usd > 0 && (
          <span className="sn-cost">${span.cost_usd.toFixed(6)}</span>
        )}
        <span
          className={`sn-dot sn-dot-${span.status}`}
          title={span.status}
        />
      </div>

      {/* ── Timeline bar ── */}
      <div className="sn-tl" style={{ marginLeft: `${indent + 1.25}rem` }}>
        <div
          className="sn-tl-bar"
          style={{ left: `${leftPct}%`, width: `${widthPct}%`, background: barColor }}
        />
      </div>

      {/* ── Children ── */}
      {expanded && hasChildren && (
        <div className="sn-children">
          {span.children!.map(child => (
            <SpanNode
              key={child.span_id}
              span={child}
              depth={depth + 1}
              totalDuration={totalDuration}
            />
          ))}
        </div>
      )}
    </div>
  );
});
