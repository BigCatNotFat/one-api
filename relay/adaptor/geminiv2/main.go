package geminiv2

import (
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/relay/meta"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	baseURL := strings.TrimSuffix(meta.BaseURL, "/")
	// Remove trailing /openai if present to avoid duplication
	baseURL = strings.TrimSuffix(baseURL, "/openai")
	// Keep the original request path (e.g., /v1/chat/completions -> /chat/completions)
	requestPath := strings.TrimPrefix(meta.RequestURLPath, "/v1")
	// Gemini OpenAI-compatible endpoint requires /openai prefix
	// e.g., https://generativelanguage.googleapis.com/v1beta/openai/chat/completions
	return fmt.Sprintf("%s/openai%s", baseURL, requestPath), nil
}
