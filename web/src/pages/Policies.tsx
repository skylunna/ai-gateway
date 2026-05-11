import { useEffect, useState } from 'react';
import { RefreshCw, Plus, Trash2, X } from 'lucide-react';
import { fetchPolicies, createPolicy, deletePolicy } from '../api/client';
import type { CreatePolicyRequest } from '../api/client';
import type { Policy } from '../types/trace';

const ACTION_BADGE: Record<string, string> = {
  block:     'bg-red-900/50    text-red-300    border border-red-700/40',
  alert:     'bg-amber-900/50  text-amber-300  border border-amber-700/40',
  downgrade: 'bg-blue-900/50   text-blue-300   border border-blue-700/40',
};
const ACTION_LABEL: Record<string, string> = {
  block: 'Block', alert: 'Alert', downgrade: 'Downgrade',
};

const EXAMPLES = [
  { label: 'Block expensive model',     expr: 'model == "gpt-4o"' },
  { label: 'Rate-limit a user (>100/h)', expr: 'user_id == "alice" && request_count > 100' },
  { label: 'Alert on high spend',       expr: 'cost_usd > 1.0' },
  { label: 'Cap token usage',           expr: 'tokens_used > 50000' },
];

/* ── Main page ────────────────────────────────────────────── */

export function Policies() {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading,  setLoading]  = useState(true);
  const [error,    setError]    = useState('');
  const [showNew,  setShowNew]  = useState(false);

  const load = () => {
    setLoading(true);
    setError('');
    fetchPolicies()
      .then(setPolicies)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete policy "${name}"?`)) return;
    try {
      await deletePolicy(id);
      setPolicies(ps => ps.filter(p => p.id !== id));
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <div className="animate-fade-in">
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Policies</h1>
          <p className="text-xs text-slate-500 mt-0.5">Gateway routing, caching and rate-limit rules</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={load} className="btn-secondary flex items-center gap-1.5 py-1.5 text-xs">
            <RefreshCw className="w-3.5 h-3.5" />
            Refresh
          </button>
          <button onClick={() => setShowNew(true)} className="btn-primary flex items-center gap-1.5 py-1.5 text-xs">
            <Plus className="w-3.5 h-3.5" />
            New Policy
          </button>
        </div>
      </div>

      {error && (
        <div className="text-red-400 bg-red-900/20 border border-red-800/40 rounded-lg px-4 py-3 mb-4 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-600 text-sm">Loading…</div>
      ) : (
        <div className="card overflow-hidden">
          {policies.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-600">
              <ShieldIcon className="w-8 h-8 mb-3 opacity-30" />
              <p className="text-sm">No policies defined</p>
              <p className="text-xs mt-1 text-slate-700">Create a policy to control LLM request behaviour</p>
            </div>
          ) : (
            policies.map((p, i) => (
              <div
                key={p.id}
                className={`flex items-center gap-4 px-5 py-3.5 hover:bg-surface-750 transition-colors
                            ${i < policies.length - 1 ? 'border-b border-surface-600' : ''}`}
              >
                <span className={`inline-block px-2 py-0.5 rounded text-[11px] font-bold uppercase
                                  tracking-wide flex-shrink-0 ${ACTION_BADGE[p.action] ?? ACTION_BADGE.block}`}>
                  {ACTION_LABEL[p.action] ?? p.action}
                </span>
                <span className="text-sm font-semibold text-slate-200 flex-shrink-0">{p.name}</span>
                {p.description && (
                  <span className="text-xs text-slate-500 flex-shrink-0">{p.description}</span>
                )}
                <span className="text-xs text-slate-600 font-mono truncate flex-1 min-w-0">
                  {p.expression}
                </span>
                {p.priority > 0 && (
                  <span className="text-[11px] text-slate-600 flex-shrink-0">priority {p.priority}</span>
                )}
                <span className={`badge flex-shrink-0 ${p.enabled ? 'badge-success' : 'badge-warning'}`}>
                  {p.enabled ? 'active' : 'inactive'}
                </span>
                <button
                  onClick={() => handleDelete(p.id, p.name)}
                  className="flex-shrink-0 p-1 text-slate-600 hover:text-red-400 transition-colors rounded"
                  title="Delete policy"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            ))
          )}
        </div>
      )}

      {showNew && (
        <NewPolicyModal
          onClose={() => setShowNew(false)}
          onCreate={p => { setPolicies(ps => [...ps, p]); setShowNew(false); }}
        />
      )}
    </div>
  );
}

/* ── New Policy Modal ─────────────────────────────────────── */

interface ModalProps {
  onClose: () => void;
  onCreate: (p: Policy) => void;
}

const EMPTY: CreatePolicyRequest = {
  name: '', expression: '', action: 'block',
  priority: 0, description: '', enabled: true,
};

function NewPolicyModal({ onClose, onCreate }: ModalProps) {
  const [form,    setForm]    = useState<CreatePolicyRequest>(EMPTY);
  const [saving,  setSaving]  = useState(false);
  const [err,     setErr]     = useState('');

  const set = <K extends keyof CreatePolicyRequest>(k: K, v: CreatePolicyRequest[K]) =>
    setForm(f => ({ ...f, [k]: v }));

  const submit = async () => {
    if (!form.name.trim())       { setErr('Name is required'); return; }
    if (!form.expression.trim()) { setErr('Expression is required'); return; }
    setSaving(true);
    setErr('');
    try {
      const p = await createPolicy(form);
      onCreate(p);
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={e => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="bg-surface-800 border border-surface-500 rounded-xl shadow-2xl w-full max-w-lg mx-4">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-surface-600">
          <h2 className="text-sm font-semibold text-slate-100">New Policy</h2>
          <button onClick={onClose} className="text-slate-500 hover:text-slate-300 transition-colors">
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <div className="px-5 py-4 space-y-4">
          {err && (
            <div className="text-red-400 bg-red-900/20 border border-red-800/40 rounded-lg px-3 py-2 text-xs">
              {err}
            </div>
          )}

          {/* Name */}
          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">
              Name <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={form.name}
              onChange={e => set('name', e.target.value)}
              placeholder="e.g. Block GPT-4o"
              className="w-full bg-surface-700 border border-surface-500 rounded-lg px-3 py-2
                         text-sm text-slate-200 placeholder-slate-600
                         focus:outline-none focus:border-primary-500 transition-colors"
            />
          </div>

          {/* Expression */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-xs font-medium text-slate-400">
                CEL Expression <span className="text-red-400">*</span>
              </label>
              <span className="text-[10px] text-slate-600">
                vars: model · user_id · tenant_id · request_count · cost_usd · tokens_used
              </span>
            </div>
            <textarea
              value={form.expression}
              onChange={e => set('expression', e.target.value)}
              rows={3}
              placeholder='model == "gpt-4o"'
              className="w-full bg-surface-700 border border-surface-500 rounded-lg px-3 py-2
                         text-sm text-slate-200 placeholder-slate-600 font-mono resize-none
                         focus:outline-none focus:border-primary-500 transition-colors"
            />
            {/* Quick-fill examples */}
            <div className="flex flex-wrap gap-1.5 mt-2">
              {EXAMPLES.map(ex => (
                <button
                  key={ex.expr}
                  onClick={() => set('expression', ex.expr)}
                  className="px-2 py-0.5 rounded text-[10px] bg-surface-700 border border-surface-500
                             text-slate-500 hover:text-slate-300 hover:border-surface-400 transition-colors"
                >
                  {ex.label}
                </button>
              ))}
            </div>
          </div>

          {/* Action + Priority row */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">Action</label>
              <select
                value={form.action}
                onChange={e => set('action', e.target.value as CreatePolicyRequest['action'])}
                className="w-full bg-surface-700 border border-surface-500 rounded-lg px-3 py-2
                           text-sm text-slate-200 focus:outline-none focus:border-primary-500 transition-colors"
              >
                <option value="block">Block</option>
                <option value="alert">Alert</option>
                <option value="downgrade">Downgrade</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">Priority</label>
              <input
                type="number"
                min={0}
                value={form.priority}
                onChange={e => set('priority', parseInt(e.target.value) || 0)}
                className="w-full bg-surface-700 border border-surface-500 rounded-lg px-3 py-2
                           text-sm text-slate-200 focus:outline-none focus:border-primary-500 transition-colors"
              />
            </div>
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">Description</label>
            <input
              type="text"
              value={form.description}
              onChange={e => set('description', e.target.value)}
              placeholder="Optional — shown in policy list"
              className="w-full bg-surface-700 border border-surface-500 rounded-lg px-3 py-2
                         text-sm text-slate-200 placeholder-slate-600
                         focus:outline-none focus:border-primary-500 transition-colors"
            />
          </div>

          {/* Enabled toggle */}
          <label className="flex items-center gap-3 cursor-pointer select-none">
            <div
              onClick={() => set('enabled', !form.enabled)}
              className={`w-9 h-5 rounded-full transition-colors ${form.enabled ? 'bg-primary-600' : 'bg-surface-600'}`}
            >
              <div className={`w-4 h-4 mt-0.5 rounded-full bg-white shadow transition-transform
                               ${form.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
            </div>
            <span className="text-xs text-slate-400">Enable immediately</span>
          </label>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 px-5 py-4 border-t border-surface-600">
          <button onClick={onClose} className="btn-secondary py-1.5 text-xs">Cancel</button>
          <button
            onClick={submit}
            disabled={saving}
            className="btn-primary py-1.5 text-xs disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saving ? 'Creating…' : 'Create Policy'}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ── Shield icon (empty state) ────────────────────────────── */

function ShieldIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round"
        d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.955 11.955 0 003 10.5c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.249-8.25-3.286z" />
    </svg>
  );
}
