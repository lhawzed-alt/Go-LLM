package optimizer

import (
	"strings"
	"unicode"
)

// systemPrompt 注入到用户消息前的简洁指令，减少输出废话
const systemPrompt = "禁止复述或转述用户请求（如“用户要我…”），禁止客套话和总结性废话，直接给出答案。"

// CleanText 清理文本中的多余空白：
// - 去除每行首尾空白
// - 合并连续空行为单个换行
// - 合并行内连续空白为单个空格（保留代码缩进场景由 collapseSpaces 控制）
func CleanText(s string) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	prevEmpty := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			if !prevEmpty && len(out) > 0 {
				out = append(out, "")
				prevEmpty = true
			}
			continue
		}
		prevEmpty = false
		out = append(out, t)
	}
	res := strings.TrimRight(strings.Join(out, "\n"), "\n")
	return res
}

// CollapseInlineSpaces 将字符串内部的连续空白压缩为单个空格（用于短字段如标题、问题）
func CollapseInlineSpaces(s string) string {
	return strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")
}

// OptimizeMessages 对 messages 做两件事：
// 1. 清理 content 中的多余空白，减少输入 token
// 2. 在最后一条 user 消息前注入简洁指令，减少输出 token
func OptimizeMessages(msgs []Message) []Message {
	if len(msgs) == 0 {
		return msgs
	}
	// 找最后一条 user 消息的位置
	lastUser := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			lastUser = i
			break
		}
	}
	for i := range msgs {
		msgs[i].Content = CleanText(msgs[i].Content)
		if i == lastUser {
			msgs[i].Content = systemPrompt + "\n" + msgs[i].Content
		}
	}
	return msgs
}

// Message 与 OpenAI 消息结构对齐（仅网关关心的字段）
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
