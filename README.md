# go_llm — AI 模型中转站

使用 Go 实现的 OpenAI 兼容 API 网关（中转站），支持多上游渠道、按模型路由、加权负载均衡、流式（SSE）转发和 API Key 鉴权。

## 功能

- ✅ OpenAI 经典接口：`POST /v1/chat/completions`、`GET /v1/models`
- ✅ 多上游渠道（OpenAI / DeepSeek / 任意兼容服务），按模型名自动路由
- ✅ 同一模型多渠道时加权随机负载均衡
- ✅ 流式响应（SSE）透传，逐块刷新
- ✅ Bearer Key 鉴权（可配置多个客户端 Key）
- ✅ 配置文件支持 `${ENV_VAR}` 引用上游 Key

## 快速开始

```bash
go mod tidy
go run ./cmd/server -config config.yaml
```

## 测试

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-relay-test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"你好"}]}'
```

流式请求在 body 中加 `"stream": true` 即可。

## 配置

见 `config.yaml`。上游 `api_key` 建议用环境变量：`${OPENAI_API_KEY}`。
