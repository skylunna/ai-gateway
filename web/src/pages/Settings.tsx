const EXAMPLE_CONFIG = `providers:
  - name: openai-prod
    base_url: "https://api.openai.com/v1"
    api_key: "\${OPENAI_API_KEY}"
    models: ["gpt-4o", "gpt-4o-mini"]
    timeout: "30s"

cache:
  enabled: true
  max_items: 5000
  default_ttl: "2h"

rate_limit:
  enabled: true
  providers:
    - name: openai-prod
      qps: 60.0
      burst: 10`;

// Simple YAML syntax highlighter
function highlight(yaml: string): string {
  return yaml
    .replace(/^(\s*)([\w_-]+)(:)/gm, '$1<span class="text-blue-400">$2</span><span class="text-slate-500">$3</span>')
    .replace(/"([^"]*)"/g, '<span class="text-emerald-400">"$1"</span>')
    .replace(/\b(true|false)\b/g, '<span class="text-amber-400">$1</span>')
    .replace(/\b(\d+\.?\d*)\b/g, '<span class="text-violet-400">$1</span>')
    .replace(/(\$\{[^}]+\})/g, '<span class="text-amber-300">$1</span>');
}

export function Settings() {
  return (
    <div className="animate-fade-in">
      {/* Page header */}
      <div className="mb-6">
        <h1 className="text-xl font-bold text-slate-100">Settings</h1>
        <p className="text-xs text-slate-500 mt-0.5">
          Gateway configuration — hot-reloaded on save
        </p>
      </div>

      {/* Config editor */}
      <div className="card overflow-hidden max-w-2xl">
        {/* Header bar */}
        <div className="flex items-center justify-between px-4 py-2.5 bg-surface-900 border-b border-surface-600">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-red-500/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/60" />
          </div>
          <span className="text-[11px] text-slate-600 font-mono">config.yaml</span>
          <span className="text-[11px] text-slate-600">read-only</span>
        </div>

        {/* Code block */}
        <pre className="px-5 py-4 text-[12px] font-mono leading-relaxed overflow-auto">
          <code
            className="text-slate-300"
            dangerouslySetInnerHTML={{ __html: highlight(EXAMPLE_CONFIG) }}
          />
        </pre>
      </div>

      {/* Help note */}
      <p className="mt-4 text-[11px] text-slate-600">
        Edit <span className="text-slate-500 font-mono">/etc/luner/config.yaml</span> to update gateway settings.
        Changes apply automatically without restarting the process.
      </p>
    </div>
  );
}
