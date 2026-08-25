# go_llm — AI 模型中转站

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

使用 Go 实现的 OpenAI 兼容 API 网关（中转站），支持多上游渠道、按模型路由、加权负载均衡、流式（SSE）转发、多用户 API Key 管理和 Token 优化。

## 功能

- ✅ OpenAI 经典接口：`POST /v1/chat/completions`、`GET /v1/models`
- ✅ 多上游渠道（OpenAI / DeepSeek / 任意兼容服务），按模型名自动路由
- ✅ 同一模型多渠道时加权随机负载均衡
- ✅ 流式响应（SSE）透传，逐块刷新
- ✅ 多用户 Bearer Key 鉴权，支持运行时动态增删 Key
- ✅ 配置文件支持 `${ENV_VAR}` 引用上游 Key
- ✅ 输入优化：自动清理消息中的多余空白，减少输入 token
- ✅ 输出优化：注入简洁指令，抑制"用户要我…"等废话，减少输出 token
- ✅ 响应缓存：非流式请求 LRU + TTL 缓存，重复请求直接命中
- ✅ 监控接口：实时查看缓存命中率

## 快速开始

```bash
go mod tidy

# 首次使用：从示例配置复制出 config.yaml，并填入真实的上游 API Key
cp config.yaml.example config.yaml   # Windows PowerShell: Copy-Item config.yaml.example config.yaml

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

Python 客户端示例见 `test.py`（注意 `base_url` 需带 `/v1` 后缀）。

## 多用户 API Key 管理

在 `config.yaml` 中预置多个客户端 Key：

```yaml
auth:
  api_keys:
    - sk-relay-test-key-123
    - sk-user-alice
    - sk-user-bob
```

也可通过管理接口运行时增删（无需重启）：

```bash
# 查看所有 Key（脱敏显示）
curl http://localhost:8080/admin/keys

# 添加新 Key
curl -X POST http://localhost:8080/admin/keys -d '{"key":"sk-user-alice"}'

# 吊销 Key
curl -X DELETE http://localhost:8080/admin/keys -d '{"key":"sk-user-alice"}'
```

> ⚠️ `/admin/keys` 目前无鉴权，生产环境请加管理员校验或仅监听内网。

## 监控

```bash
# 缓存命中率
curl http://localhost:8080/stats
# 健康检查
curl http://localhost:8080/healthz
```

## 项目结构

```
cmd/server/main.go          # 入口与路由注册
internal/config/config.go   # YAML 配置加载（支持环境变量）
internal/router/router.go   # 模型路由 + 加权负载均衡
internal/proxy/proxy.go     # 核心转发、SSE 流式透传、缓存接入
internal/auth/auth.go       # 多用户 Key Store 与鉴权中间件
internal/optimizer/optimizer.go # 消息清理与简洁指令注入
internal/cache/cache.go     # LRU + TTL 响应缓存
```

## 配置

仓库中只包含脱敏的 `config.yaml.example`，首次使用需复制为 `config.yaml` 并填入真实 Key：

```bash
cp config.yaml.example config.yaml
```

> `config.yaml` 已被 `.gitignore` 排除，不会被提交到 Git，可放心填写真实 API Key。

上游 `api_key` 建议用环境变量引用：`${OPENAI_API_KEY}`。

## 许可证

本项目基于 [Apache License 2.0](LICENSE) 开源。
