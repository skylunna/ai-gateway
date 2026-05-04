import { memo, useState } from 'react';
import { ChevronRight } from 'lucide-react';
import type { SpanNode as SpanNodeData } from '../types/trace';

const TYPE_BADGE: Record<string, string> = {
  llm:       'bg-blue-100 text-blue-800',
  tool:      'bg-green-100 text-green-800',
  retrieval: 'bg-purple-100 text-purple-800',
  custom:    'bg-gray-100 text-gray-700',
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

  const safe     = totalDuration > 0 ? totalDuration : 1;
  const leftPct  = (span.relative_start_ms / safe) * 100;
  const widthPct = Math.max((span.duration_ms / safe) * 100, 0.5);

  const badgeClass = TYPE_BADGE[span.span_type] ?? TYPE_BADGE.custom;
  const barColor   = TYPE_COLOR[span.span_type] ?? '#9ca3af';
  const indentRem  = 1 + depth * 1.25;

  return (
    <div className="border-b border-gray-100 last:border-b-0">
      {/* Info row */}
      <div
        className="flex items-center gap-1.5 pr-4 py-1.5 min-h-[34px] hover:bg-gray-50 transition-colors duration-100"
        style={{ paddingLeft: `${indentRem}rem` }}
      >
        {/* Expand / collapse */}
        <button
          className={`w-4 flex-shrink-0 flex items-center justify-center bg-transparent border-0 p-0
                      text-gray-400 hover:text-gray-600 transition-transform duration-200
                      ${hasChildren ? 'cursor-pointer' : 'cursor-default opacity-0'}`}
          style={{ transform: hasChildren && expanded ? 'rotate(90deg)' : 'rotate(0deg)' }}
          onClick={() => hasChildren && setExpanded(e => !e)}
          tabIndex={hasChildren ? 0 : -1}
          aria-label={expanded ? 'collapse' : 'expand'}
        >
          <ChevronRight className="w-3.5 h-3.5" />
        </button>

        {/* Type badge */}
        <span className={`inline-block px-1.5 py-0.5 rounded-full text-[10px] font-semibold
                          uppercase tracking-wide flex-shrink-0 ${badgeClass}`}>
          {span.span_type}
        </span>

        {/* Name */}
        <span
          className="font-medium flex-1 min-w-0 overflow-hidden text-ellipsis whitespace-nowrap text-sm text-gray-800"
          title={span.name}
        >
          {span.name}
        </span>

        {/* Model */}
        {span.model && (
          <span className="text-[11px] text-gray-400 flex-shrink-0">{span.model}</span>
        )}

        {/* Duration */}
        <span className="text-[11px] text-gray-500 flex-shrink-0 whitespace-nowrap">
          {span.duration_ms}ms
        </span>

        {/* Cost */}
        {span.cost_usd != null && span.cost_usd > 0 && (
          <span className="text-[11px] text-primary-600 flex-shrink-0 whitespace-nowrap">
            ${span.cost_usd.toFixed(6)}
          </span>
        )}

        {/* Status dot */}
        <span
          title={span.status}
          className={`w-2 h-2 rounded-full flex-shrink-0 ${
            span.status === 'success' ? 'bg-green-500' :
            span.status === 'timeout' ? 'bg-yellow-500' : 'bg-red-500'
          }`}
        />
      </div>

      {/* Timeline bar */}
      <div
        className="relative h-1.5 bg-gray-100 mr-4 mb-1.5 rounded-full"
        style={{ marginLeft: `${indentRem + 1.5}rem` }}
      >
        <div
          className="absolute top-0 bottom-0 rounded-full opacity-75"
          style={{
            left: `${leftPct}%`,
            width: `${widthPct}%`,
            background: barColor,
            minWidth: '2px',
          }}
        />
      </div>

      {/* Children */}
      {expanded && hasChildren && (
        <div>
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
