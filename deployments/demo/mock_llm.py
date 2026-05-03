#!/usr/bin/env python3
"""Mock OpenAI-compatible LLM server for the luner demo."""
import http.server
import json
import sys
import time
from socketserver import ThreadingMixIn


class ThreadedHTTPServer(ThreadingMixIn, http.server.HTTPServer):
    daemon_threads = True
    allow_reuse_address = True


class MockLLMHandler(http.server.BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        sys.stderr.write(f"[mock-llm] {fmt % args}\n")

    def do_GET(self):
        if self.path in ("/health", "/v1/models"):
            self._json(200, {"status": "ok", "models": ["gpt-4o-mini", "gpt-4o"]})
        else:
            self._json(404, {"error": "not found"})

    def do_POST(self):
        try:
            length = int(self.headers.get("Content-Length", 0))
            body = json.loads(self.rfile.read(length)) if length else {}
            model = body.get("model", "gpt-4o-mini")

            # Simulate a slight latency so duration_ms is non-zero
            time.sleep(0.1)

            self._json(200, {
                "id": "chatcmpl-demo",
                "object": "chat.completion",
                "created": int(time.time()),
                "model": model,
                "choices": [{
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": "This is a mock response. In production this would be a real LLM reply."
                    },
                    "finish_reason": "stop"
                }],
                "usage": {
                    "prompt_tokens": 15,
                    "completion_tokens": 20,
                    "total_tokens": 35
                }
            })
        except Exception as e:
            self._json(500, {"error": str(e)})

    def _json(self, status, data):
        body = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    server = ThreadedHTTPServer(("0.0.0.0", 19999), MockLLMHandler)
    print("Mock LLM listening on :19999", file=sys.stderr, flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        server.shutdown()
