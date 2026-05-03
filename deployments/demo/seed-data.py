#!/usr/bin/env python3
"""Generate sample trace data for the luner demo dashboard."""
import json
import os
import sys
import time
import urllib.request
import urllib.error

LUNER_URL = os.getenv("LUNER_URL", "http://localhost:8080")


def wait_for_luner(max_tries=30):
    for i in range(max_tries):
        try:
            with urllib.request.urlopen(f"{LUNER_URL}/api/health", timeout=2) as r:
                if r.status == 200:
                    print("luner is ready", flush=True)
                    return True
        except Exception:
            pass
        print(f"  waiting for luner... ({i+1}/{max_tries})", flush=True)
        time.sleep(2)
    print("luner did not become ready in time", file=sys.stderr)
    return False


def chat(agent, user, tenant, content):
    payload = json.dumps({
        "model": "gpt-4o-mini",
        "messages": [{"role": "user", "content": content}]
    }).encode()
    req = urllib.request.Request(
        f"{LUNER_URL}/v1/chat/completions",
        data=payload,
        headers={
            "Content-Type": "application/json",
            "X-Luner-Agent": agent,
            "X-Luner-User": user,
            "X-Luner-Tenant": tenant,
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status == 200
    except urllib.error.HTTPError as e:
        print(f"  HTTP {e.code}: {e.read()[:120]}", file=sys.stderr)
        return False
    except Exception as e:
        print(f"  error: {e}", file=sys.stderr)
        return False


def create_policy(name, expression, action, description=""):
    payload = json.dumps({
        "name": name,
        "expression": expression,
        "action": action,
        "description": description,
        "priority": 100,
        "enabled": False,
    }).encode()
    req = urllib.request.Request(
        f"{LUNER_URL}/api/policies",
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as r:
            return r.status in (200, 201)
    except Exception as e:
        print(f"  policy create failed: {e}", file=sys.stderr)
        return False


def main():
    if not wait_for_luner():
        sys.exit(1)

    print("\nSeeding demo traces...", flush=True)

    scenarios = [
        ("code-reviewer",   "alice", "acme-corp", [
            "Review PR #101: add authentication middleware",
            "Review PR #102: optimize database queries",
            "Review PR #103: update API rate limiting",
        ]),
        ("doc-writer",      "bob",   "acme-corp", [
            "Write API reference for /v1/chat/completions",
            "Write deployment guide for Docker Compose",
        ]),
        ("test-generator",  "alice", "acme-corp", [
            "Generate unit tests for the auth service",
            "Generate integration tests for the payment flow",
        ]),
        ("data-analyst",    "carol", "beta-inc",  [
            "Summarize Q4 sales report",
        ]),
    ]

    total, ok = 0, 0
    for agent, user, tenant, tasks in scenarios:
        for task in tasks:
            print(f"  {agent} / {user}: {task[:50]}", flush=True)
            if chat(agent, user, tenant, task):
                ok += 1
            total += 1
            time.sleep(0.3)

    print(f"\n{ok}/{total} traces seeded", flush=True)

    print("\nCreating demo policies...", flush=True)
    policies = [
        ("Block GPT-4o in demo",
         'model == "gpt-4o"',
         "block",
         "Prevent expensive model usage (disabled by default)"),
        ("Alert on high spend",
         "cost_usd > 1.0",
         "alert",
         "Alert when a user exceeds $1 total spend (disabled by default)"),
    ]
    for name, expr, action, desc in policies:
        if create_policy(name, expr, action, desc):
            print(f"  created: {name}", flush=True)

    print("\nDemo data ready. Open http://localhost:8080 to explore.", flush=True)


if __name__ == "__main__":
    main()
