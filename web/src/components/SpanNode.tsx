import { memo, useState } from 'react';
import { ChevronRight } from 'lucide-react';
import type { SpanNode as SpanNodeData } from '../types/trace';

const TYPE_BADGE: Record<string, string> = {
  llm:       'bg-blue-900/60   text-blue-300   border border-blue-700/40',
  tool:      'bg-green-900/60  text-green-300  border border-green-700/40',
  retrieval: 'bg-violet-900/60 text-violet-300 border border-violet-700/40',
  custom:    'bg-slate-700/50  text-slate-400  border border-slate-600/40',
};

const TYPE_COLOR: Record<string, string> = {
  llm:       '#3b82f6',
  tool:      '#10b981',
  retrieval: '#8b5cf6',
  custom:    '#64748b',
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
  const barColor   = TYPE_COLOR[span.span_type] ?? '#64748b';
  const indentRem  = 1 + depth * 1.25;

  const hasTokens  = span.prompt_tokens != null && span.prompt_tokens > 0;

  return (
    <div className="border-b border-surface-600/60 last:border-b-0">
      {/* ── Info row ── */}
      <div
        className="flex items-center gap-2 pr-4 py-1.5 min-h-[36px] hover:bg-surface-750/50 transition-colors duration-100 group"
        style={{ paddingLeft: `${indentRem}rem` }}
      >
        {/* Expand toggle */}
        <button
          className={`w-4 h-4 flex items-center justify-center flex-shrink-0 rounded
                      text-slate-600 hover:text-slate-400 transition-all duration-200
                      ${hasChildren ? 'cursor-pointer' : 'cursor-default opacity-0 pointer-events-none'}`}
          style={{ transform: hasChildren && expanded ? 'rotate(90deg)' : 'rotate(0deg)' }}
          onClick={() => hasChildren && setExpanded(e => !e)}
          tabIndex={hasChildren ? 0 : -1}
          aria-label={expanded ? 'collapse' : 'expand'}
        >
          <ChevronRight className="w-3.5 h-3.5" />
        </button>

        {/* Type badge */}
        <span className={`inline-block px-1.5 py-px rounded text-[10px] font-bold
                          uppercase tracking-wide flex-shrink-0 ${badgeClass}`}>
          {span.span_type}
        </span>

        {/* Name */}
        <span
          className="font-medium flex-1 min-w-0 overflow-hidden text-ellipsis whitespace-nowrap text-[13px] text-slate-200"
          title={span.name}
        >
          {span.name}
        </span>

        {/* Right-side metadata */}
        <div className="flex items-center gap-3 flex-shrink-0">
          {/* Model */}
          {span.model && (
            <span className="text-[11px] text-slate-500">{span.model}</span>
          )}

          {/* Duration */}
          <span className="text-[11px] text-slate-400 tabular-nums whitespace-nowrap">
            {span.duration_ms}ms
          </span>

          {/* Cost */}
          {span.cost_usd != null && span.cost_usd > 0 && (
            <span className="text-[11px] text-primary-400 tabular-nums whitespace-nowrap font-medium">
              ${span.cost_usd.toFixed(4)}
            </span>
          )}

          {/* Token counts: ●prompt + ●completion */}
          {hasTokens && (
            <span className="hidden sm:flex items-center gap-1 text-[10px] text-slate-500 tabular-nums whitespace-nowrap">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-500 flex-shrink-0" />
              {span.prompt_tokens}
              <span className="text-slate-600">+</span>
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
              {span.completion_tokens ?? 0}
            </span>
          )}

          {/* Status dot */}
          <span
            title={span.status}
            className={`w-2 h-2 rounded-full flex-shrink-0 ${
              span.status === 'success' ? 'bg-emerald-500' :
              span.status === 'timeout' ? 'bg-amber-500'   : 'bg-red-500'
            }`}
          />
        </div>
      </div>

      {/* ── Timeline bar ── */}
      <div
        className="relative h-[3px] bg-surface-600/50 mb-2 rounded-full mx-4"
        style={{ marginLeft: `${indentRem + 1.5}rem` }}
      >
        <div
          className="absolute top-0 bottom-0 rounded-full opacity-80"
          style={{ left:`${leftPct}%`, width:`${widthPct}%`, background:barColor, minWidth:'3px' }}
        />
      </div>

      {/* ── Children ── */}
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
