import http.server
import json
import sys
from socketserver import ThreadingMixIn

class ThreadedHTTPServer(ThreadingMixIn, http.server.HTTPServer):
    daemon_threads = True

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass  # 静默
    
    def do_POST(self):
        try:
            # 读取请求体
            content_length = int(self.headers.get('Content-Length', 0))
            if content_length > 0:
                self.rfile.read(content_length)
            
            # 返回 Mock 响应
            response = {
                "id": "chatcmpl-mock",
                "object": "chat.completion",
                "created": 1234567890,
                "model": "gpt-4o-mini",
                "choices": [{
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": "This is a mock response from luner demo."
                    },
                    "finish_reason": "stop"
                }],
                "usage": {
                    "prompt_tokens": 15,
                    "completion_tokens": 10,
                    "total_tokens": 25
                }
            }
            
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
            
        except Exception as e:
            print(f"Mock error: {e}", file=sys.stderr)
            self.send_response(500)
            self.end_headers()

if __name__ == "__main__":
    server = ThreadedHTTPServer(("127.0.0.1", 19999), Handler)
    print("Mock LLM listening on :19999", file=sys.stderr, flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
