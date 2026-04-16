package proxy

/*
	核心代理逻辑（路由匹配、请求转发、流式透传）
*/
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/skylunna/ai-gateway/internal/config"
	"github.com/skylunna/ai-gateway/internal/metrics"
)

type Handler struct {
	modelMap   map[string]*config.ProviderConfig // 模型 -> 厂商配置: 模型路由表: key = 模型名, value = 对应哪个厂商
	httpClient *http.Client                      // 转发用的 HTTP 客户端: 发请求给AI厂商
	logger     *slog.Logger                      // 日志:
}

/*
构造函数

把「模型名字」和「它属于哪个厂商」绑定成一张映射表，让后续请求能快速查表路由

传入配置 + 日志 -> 初始化并返回一个代理处理器

"qwen-turbo"      →  阿里云配置
"deepseek-chat"   →  深度求索配置
"gpt-3.5-turbo"   →  OpenAI配置
*/
func NewHandler(cfg *config.Config, logger *slog.Logger) *Handler {
	// 创建一个空的路由表
	// m = 模型路由字典
	// key: string 模型名 (例如 "qwen-turbo")
	// value: *config.ProviderConfig 指向这个模型对应的厂商配置 (地址、密钥、超时等)
	m := make(map[string]*config.ProviderConfig)
	// 遍历所有厂商
	for i := range cfg.Providers {
		// 拿到当前厂商配置的指针 (不复制, 节省内存)
		p := &cfg.Providers[i]
		// 遍历厂商下的所有模型
		for _, mod := range p.Models {
			m[mod] = p
		}
	}
	/*

		m = {
		  "qwen-turbo":    &{name: "aliyun", base_url: "...", api_key: "..."},
		  "qwen-plus":     &{name: "aliyun", base_url: "...", api_key: "..."},
		  "gpt-3.5-turbo": &{name: "openai", base_url: "...", api_key: "..."},
		  "gpt-4":         &{name: "openai", base_url: "...", api_key: "..."},
		}
	*/
	return &Handler{
		modelMap:   m,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

/*
ServeHTTP 是网关的入口总控函数，所有用户请求都从这里进入
它负责：校验请求 → 解析模型 → 路由转发 → 透传响应 → 上报监控是标准的 HTTP 处理器接口实现。
这是 Go 官方 http.Handler 接口的必须实现方法

	w: 用来写响应给用户
	r: 用户发来的请求信息 (路径、头、body等)
	h: 当前处理器 (含路由表、客户端、日志)
*/
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 仅处理 OpenAI 兼容端点
	if r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	/*
		2. 读取并校验 Body（必须缓冲，用于后续转发 & 提取 model）

		 	把用户发送的JSON完整读成字节数组
			1. 解析里面的 model
			2. 转发给上游厂商
	*/
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("read request body", "err", err)
		h.writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	// 关闭请求流, 避免内存泄露
	defer r.Body.Close()

	var reqPayload map[string]any
	if err := json.Unmarshal(bodyBytes, &reqPayload); err != nil {
		h.writeError(w, "invalid json", http.StatusBadRequest)
		return
	}

	model, _ := reqPayload["model"].(string)
	if model == "" {
		h.writeError(w, "model is required", http.StatusBadRequest)
		metrics.RequestTotal.WithLabelValues("unknown", "unknown", "400").Inc()
		return
	}

	// 3. 路由匹配
	prov, ok := h.modelMap[model]
	if !ok {
		h.writeError(w, fmt.Sprintf("unsupported model: %s", model), http.StatusBadRequest)
		metrics.RequestTotal.WithLabelValues(model, "unknown", "400").Inc()
		return
	}

	// 4. 构造上游请求（带超时控制）
	timeout := prov.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	targetURL := prov.BaseURL + r.URL.Path
	targetURL = strings.ReplaceAll(targetURL, "/v1/v1", "/v1")

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		h.logger.Error("create upstream request", "err", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 复制原 Header 并注入 API Key
	upstreamReq.Header = r.Header.Clone()
	upstreamReq.Header.Set("Authorization", "Bearer "+prov.APIKey)
	upstreamReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
	upstreamReq.ContentLength = int64(len(bodyBytes))

	// 5. 发起请求 & 记录指标
	start := time.Now()
	resp, err := h.httpClient.Do(upstreamReq)
	duration := time.Since(start).Seconds()

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	// 请求总数 +1
	// 按模型、厂商、状态码维度统计
	metrics.RequestTotal.WithLabelValues(model, prov.Name, fmt.Sprintf("%d", statusCode)).Inc()
	// 记录耗时分布
	metrics.RequestDuration.WithLabelValues(model, prov.Name).Observe(duration)

	if err != nil {
		h.logger.Error("upstream request failed", "err", err)
		http.Error(w, `{"error":"upstream timeout or network error"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 6. 透传响应（完美支持流式 SSE）
	// 原样返回厂商的响应头
	// 原样返回状态码
	// io.Copy 流式回写给用户
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}
