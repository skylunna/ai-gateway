package trace

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// Context holds Agent metadata extracted from an HTTP request's headers.
type Context struct {
	// W3C Trace Context (from traceparent header)
	TraceID      string
	SpanID       string
	ParentSpanID string
	Sampled      bool

	// Luner business context (from X-Luner-* headers)
	AgentName    string
	AgentVersion string
	SessionID    string
	UserID       string
	TenantID     string
	Environment  string
	Tags         map[string]string
}

// ExtractContext parses W3C traceparent and X-Luner-* headers.
// It always returns a non-nil Context; missing fields are empty strings.
func ExtractContext(headers http.Header) *Context {
	ctx := &Context{Tags: make(map[string]string)}

	// traceparent: 00-{trace_id}-{span_id}-{flags}
	if tp := headers.Get("traceparent"); tp != "" {
		parts := strings.Split(tp, "-")
		if len(parts) == 4 {
			ctx.TraceID = parts[1]
			ctx.SpanID = parts[2]
			ctx.Sampled = parts[3] == "01"
		}
	}

	ctx.AgentName = headers.Get("X-Luner-Agent")
	ctx.AgentVersion = headers.Get("X-Luner-Agent-Version")
	ctx.SessionID = headers.Get("X-Luner-Session")
	ctx.UserID = headers.Get("X-Luner-User")
	ctx.TenantID = headers.Get("X-Luner-Tenant")
	ctx.Environment = headers.Get("X-Luner-Env")
	ctx.ParentSpanID = headers.Get("X-Luner-Parent-Span")

	// X-Luner-Tags: k1=v1,k2=v2
	if raw := headers.Get("X-Luner-Tags"); raw != "" {
		for _, pair := range strings.Split(raw, ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				ctx.Tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	return ctx
}

// IsEmpty reports whether the context carries no meaningful identity.
func (c *Context) IsEmpty() bool {
	return c.TraceID == "" && c.AgentName == "" && c.UserID == ""
}

// GenerateTraceID sets TraceID to a fresh 32-char hex string if it is empty.
// Idempotent: calling it again on a non-empty TraceID is a no-op.
func (c *Context) GenerateTraceID() {
	if c.TraceID == "" {
		c.TraceID = newHexID(32)
	}
}

// GenerateSpanID sets SpanID to a fresh 16-char hex string if it is empty.
func (c *Context) GenerateSpanID() {
	if c.SpanID == "" {
		c.SpanID = newHexID(16)
	}
}

// newHexID returns a cryptographically random hex string of `chars` characters.
func newHexID(chars int) string {
	b := make([]byte, chars/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
