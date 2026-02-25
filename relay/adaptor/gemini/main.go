package gemini

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/render"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/image"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"

	"github.com/gin-gonic/gin"
)

// https://ai.google.dev/docs/gemini_api_overview?hl=zh-cn

const (
	VisionMaxImageNum = 16
)

var mimeTypeMap = map[string]string{
	"json_object": "application/json",
	"text":        "text/plain",
}

func geminiFinishReason2OpenAI(reason string) string {
	switch strings.ToUpper(reason) {
	case "", "FINISH_REASON_UNSPECIFIED", "STOP":
		return constant.StopFinishReason
	case "MAX_TOKENS":
		return "length"
	default:
		// SAFETY / RECITATION / OTHER ... keep as-is for transparency
		return strings.ToLower(reason)
	}
}

// geminiUnsupportedSchemaKeys contains JSON Schema fields that Gemini does not support.
// Gemini only supports a subset of OpenAPI Schema, not full JSON Schema.
var geminiUnsupportedSchemaKeys = map[string]bool{
	"additionalProperties":   true,
	"unevaluatedProperties":  true,
	"$ref":                   true,
	"$schema":                true,
	"$id":                    true,
	"$defs":                  true,
	"definitions":            true,
	"patternProperties":      true,
	"propertyNames":          true,
	"unevaluatedItems":       true,
	"contains":               true,
	"minContains":            true,
	"maxContains":            true,
	"if":                     true,
	"then":                   true,
	"else":                   true,
	"allOf":                  true,
	"anyOf":                  true,
	"oneOf":                  true,
	"not":                    true,
	"dependentSchemas":       true,
	"dependentRequired":      true,
	"const":                  true,
	"contentEncoding":        true,
	"contentMediaType":       true,
	"contentSchema":          true,
	"deprecated":             true,
	"readOnly":               true,
	"writeOnly":              true,
	"examples":               true,
	"default":                true,
	"$comment":               true,
	"$vocabulary":            true,
	"$anchor":                true,
	"$dynamicRef":            true,
	"$dynamicAnchor":         true,
	"minLength":              true,
	"maxLength":              true,
	"pattern":                true,
	"minimum":                true,
	"maximum":                true,
	"exclusiveMinimum":       true,
	"exclusiveMaximum":       true,
	"multipleOf":             true,
	"minItems":               true,
	"maxItems":               true,
	"uniqueItems":            true,
	"minProperties":          true,
	"maxProperties":          true,
}

func sanitizeGeminiFunctionParameters(parameters any) any {
	// Gemini function_declarations' parameters is NOT full JSON Schema.
	// A common incompatibility is "additionalProperties" (OpenAI-style JSON Schema).
	// We recursively drop unsupported keys to avoid Gemini 400 errors.
	switch v := parameters.(type) {
	case map[string]any:
		clean := make(map[string]any, len(v))
		for k, vv := range v {
			// drop keys that Gemini rejects
			if geminiUnsupportedSchemaKeys[k] {
				continue
			}
			clean[k] = sanitizeGeminiFunctionParameters(vv)
		}
		return clean
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, sanitizeGeminiFunctionParameters(item))
		}
		return out
	default:
		return parameters
	}
}

func convertToolChoiceToToolConfig(toolChoice any) *ToolConfig {
	if toolChoice == nil {
		return nil
	}
	// OpenAI tool_choice can be:
	// - "auto" | "none" | "required"
	// - {"type":"function","function":{"name":"xxx"}}
	cfg := &FunctionCallingConfig{}

	if tcStr, ok := toolChoice.(string); ok {
		switch tcStr {
		case "none":
			cfg.Mode = "NONE"
		case "required":
			cfg.Mode = "ANY"
		default: // "auto"
			cfg.Mode = "AUTO"
		}
		return &ToolConfig{FunctionCallingConfig: cfg}
	}
	if tcMap, ok := toolChoice.(map[string]any); ok {
		// example: {"type":"function","function":{"name":"get_weather"}}
		if t, ok := tcMap["type"].(string); ok && t == "function" {
			if fn, ok := tcMap["function"].(map[string]any); ok {
				if name, ok := fn["name"].(string); ok && name != "" {
					cfg.Mode = "ANY"
					cfg.AllowedFunctionNames = []string{name}
					return &ToolConfig{FunctionCallingConfig: cfg}
				}
			}
		}
	}
	return nil
}

func parseToolResultContentToGeminiResponse(content any) any {
	// Gemini expects `functionResponse.response` to be an object.
	// Try to parse JSON string; if not JSON, wrap into {"content": "..."}.
	switch v := content.(type) {
	case string:
		var obj any
		if err := json.Unmarshal([]byte(v), &obj); err == nil {
			return obj
		}
		return map[string]any{"content": v}
	default:
		return map[string]any{"content": fmt.Sprintf("%v", content)}
	}
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func ConvertRequest(textRequest model.GeneralOpenAIRequest) (*ChatRequest, error) {
	geminiRequest := ChatRequest{
		Contents: make([]ChatContent, 0, len(textRequest.Messages)),
		SafetySettings: []ChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_CIVIC_INTEGRITY",
				Threshold: config.GeminiSafetySetting,
			},
		},
		GenerationConfig: ChatGenerationConfig{
			Temperature:     textRequest.Temperature,
			TopP:            textRequest.TopP,
			MaxOutputTokens: textRequest.MaxTokens,
		},
	}

	// Handle extra_body parameters for Gemini-specific features
	if textRequest.ExtraBody != nil {
		if googleConfig, ok := textRequest.ExtraBody["google"].(map[string]interface{}); ok {
			if thinkingConfig, ok := googleConfig["thinking_config"].(map[string]interface{}); ok {
				geminiRequest.GenerationConfig.ThinkingConfig = thinkingConfig
			}
		}
	}
	if textRequest.ResponseFormat != nil {
		if mimeType, ok := mimeTypeMap[textRequest.ResponseFormat.Type]; ok {
			geminiRequest.GenerationConfig.ResponseMimeType = mimeType
		}
		if textRequest.ResponseFormat.JsonSchema != nil {
			geminiRequest.GenerationConfig.ResponseSchema = textRequest.ResponseFormat.JsonSchema.Schema
			geminiRequest.GenerationConfig.ResponseMimeType = mimeTypeMap["json_object"]
		}
	}
	if textRequest.Tools != nil {
		functions := make([]model.Function, 0, len(textRequest.Tools))
		for _, tool := range textRequest.Tools {
			fn := tool.Function
			fn.Parameters = sanitizeGeminiFunctionParameters(fn.Parameters)
			functions = append(functions, fn)
		}
		geminiRequest.Tools = []ChatTools{
			{
				FunctionDeclarations: functions,
			},
		}
	} else if textRequest.Functions != nil {
		// Sanitize legacy functions field as well
		sanitizedFunctions := sanitizeGeminiFunctionParameters(textRequest.Functions)
		geminiRequest.Tools = []ChatTools{
			{
				FunctionDeclarations: sanitizedFunctions,
			},
		}
	}

	// tool_choice -> tool_config (best-effort)
	geminiRequest.ToolConfig = convertToolChoiceToToolConfig(textRequest.ToolChoice)

	// Build a lookup table: tool_call_id -> function name
	toolCallNameByID := make(map[string]string)
	for _, m := range textRequest.Messages {
		for _, tc := range m.ToolCalls {
			if tc.Id != "" && tc.Function.Name != "" {
				toolCallNameByID[tc.Id] = tc.Function.Name
			}
		}
	}

	shouldAddDummyModelMessage := false
	for _, message := range textRequest.Messages {
		// OpenAI tool result message -> Gemini functionResponse part
		if message.Role == "tool" {
			name := toolCallNameByID[message.ToolCallId]
			if name == "" && message.Name != nil {
				name = *message.Name
			}
			if name == "" {
				return nil, errors.New("gemini: tool message missing tool_call_id mapping to function name")
			}
			geminiRequest.Contents = append(geminiRequest.Contents, ChatContent{
				Role: "user",
				Parts: []Part{
					{
						FunctionResp: &FunctionResponse{
							FunctionName: name,
							Response:     parseToolResultContentToGeminiResponse(message.Content),
						},
					},
				},
			})
			continue
		}

		content := ChatContent{
			Role: message.Role,
			Parts: []Part{
				{
					Text: message.StringContent(),
				},
			},
		}
		openaiContent := message.ParseContent()
		var parts []Part
		imageNum := 0
		for _, part := range openaiContent {
			if part.Type == model.ContentTypeText {
				// Skip empty text parts to avoid Gemini's "required oneof field 'data' must have one initialized field" error
				// This happens when assistant message has tool_calls but empty/null content
				if part.Text == "" {
					continue
				}
				parts = append(parts, Part{
					Text: part.Text,
				})
			} else if part.Type == model.ContentTypeImageURL {
				imageNum += 1
				if imageNum > VisionMaxImageNum {
					continue
				}
				mimeType, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				parts = append(parts, Part{
					InlineData: &InlineData{
						MimeType: mimeType,
						Data:     data,
					},
				})
			}
		}

		// Preserve OpenAI tool_calls history (assistant -> model.functionCall parts)
		// Also check for "model" role since some clients may directly use Gemini's role name
		if (message.Role == "assistant" || message.Role == "model") && len(message.ToolCalls) > 0 {
			isFirstFunctionCall := true
			for _, tc := range message.ToolCalls {
				var args any
				switch v := tc.Function.Arguments.(type) {
				case string:
					_ = json.Unmarshal([]byte(v), &args)
				default:
					args = v
				}
				funcCall := &FunctionCall{
					FunctionName: tc.Function.Name,
					Arguments:    args,
				}
				part := Part{
					FunctionCall: funcCall,
				}
				if tc.ExtraContent != nil {
					if google, ok := tc.ExtraContent["google"].(map[string]interface{}); ok {
						if sig, ok := google["thought_signature"].(string); ok && sig != "" {
							part.ThoughtSignature = sig
						}
					}
				}
				// Gemini 3 requires thought_signature on the first functionCall part per step.
				// When downstream clients don't preserve extra_content, use a dummy signature
				// to bypass validation (officially documented workaround).
				if part.ThoughtSignature == "" && isFirstFunctionCall {
					part.ThoughtSignature = "context_engineering_is_the_way_to_go"
				}
				isFirstFunctionCall = false
				parts = append(parts, part)
			}
		}
		content.Parts = parts

		// Skip messages with empty parts to avoid Gemini's "required oneof field 'data' must have one initialized field" error
		// This can happen when a message has no text content and no tool_calls
		if len(content.Parts) == 0 {
			continue
		}

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		// Converting system prompt to prompt from user for the same reason
		if content.Role == "system" {
			if IsModelSupportSystemInstruction(textRequest.Model) {
				geminiRequest.SystemInstruction = &content
				geminiRequest.SystemInstruction.Role = ""
				continue
			} else {
				shouldAddDummyModelMessage = true
				content.Role = "user"
			}
		}

		geminiRequest.Contents = append(geminiRequest.Contents, content)

		// If a system message is the last message, we need to add a dummy model message to make gemini happy
		if shouldAddDummyModelMessage {
			geminiRequest.Contents = append(geminiRequest.Contents, ChatContent{
				Role: "model",
				Parts: []Part{
					{
						Text: "Okay",
					},
				},
			})
			shouldAddDummyModelMessage = false
		}
	}

	return &geminiRequest, nil
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) *BatchEmbeddingRequest {
	inputs := request.ParseInput()
	requests := make([]EmbeddingRequest, len(inputs))
	model := fmt.Sprintf("models/%s", request.Model)

	for i, input := range inputs {
		requests[i] = EmbeddingRequest{
			Model: model,
			Content: ChatContent{
				Parts: []Part{
					{
						Text: input,
					},
				},
			},
		}
	}

	return &BatchEmbeddingRequest{
		Requests: requests,
	}
}

// UsageMetadata contains token usage information returned by Gemini API
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type ChatResponse struct {
	Candidates     []ChatCandidate    `json:"candidates"`
	PromptFeedback ChatPromptFeedback `json:"promptFeedback"`
	UsageMetadata  *UsageMetadata     `json:"usageMetadata,omitempty"`
}

func (g *ChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	if len(g.Candidates) > 0 && len(g.Candidates[0].Content.Parts) > 0 {
		return g.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

type ChatCandidate struct {
	Content       ChatContent        `json:"content"`
	FinishReason  string             `json:"finishReason"`
	Index         int64              `json:"index"`
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

type ChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type ChatPromptFeedback struct {
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

func getToolCalls(candidate *ChatCandidate) []model.Tool {
	var toolCalls []model.Tool

	for _, item := range candidate.Content.Parts {
		if item.FunctionCall == nil {
			continue
		}
		argsBytes, err := json.Marshal(item.FunctionCall.Arguments)
		if err != nil {
			logger.FatalLog("getToolCalls failed: " + err.Error())
			continue
		}
		tool := model.Tool{
			Id:   fmt.Sprintf("call_%s", random.GetUUID()),
			Type: "function",
			Function: model.Function{
				Arguments: string(argsBytes),
				Name:      item.FunctionCall.FunctionName,
			},
		}
		// Gemini 3: preserve thought_signature for tool calls (required when thinking is enabled)
		// thought_signature is at Part level, not inside FunctionCall
		thoughtSig := item.ThoughtSignature
		// Debug: log the thought_signature extraction
		if config.DebugEnabled {
			partBytes, _ := json.Marshal(item)
			logger.SysLog(fmt.Sprintf("Gemini getToolCalls Part: %s, ThoughtSig: %s", string(partBytes), thoughtSig))
		}
		if thoughtSig != "" {
			tool.ExtraContent = map[string]interface{}{
				"google": map[string]interface{}{
					"thought_signature": thoughtSig,
				},
			}
		}
		toolCalls = append(toolCalls, tool)
	}
	return toolCalls
}

func responseGeminiChat2OpenAI(response *ChatResponse) *openai.TextResponse {
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: make([]openai.TextResponseChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := openai.TextResponseChoice{
			Index: i,
			Message: model.Message{
				Role: "assistant",
			},
			FinishReason: constant.StopFinishReason,
		}
		if len(candidate.Content.Parts) > 0 {
			if len(getToolCalls(&candidate)) > 0 {
				choice.Message.ToolCalls = getToolCalls(&candidate)
				choice.FinishReason = "tool_calls"
			} else {
				var contentBuilder strings.Builder
				var reasoningBuilder strings.Builder
				for _, part := range candidate.Content.Parts {
					if part.Thought != nil && *part.Thought {
						if reasoningBuilder.Len() > 0 {
							reasoningBuilder.WriteString("\n")
						}
						reasoningBuilder.WriteString(part.Text)
					} else {
						if contentBuilder.Len() > 0 {
							contentBuilder.WriteString("\n")
						}
						contentBuilder.WriteString(part.Text)
					}
				}
				choice.Message.Content = contentBuilder.String()
				if reasoningBuilder.Len() > 0 {
					choice.Message.ReasoningContent = reasoningBuilder.String()
				}
			}
		} else {
			choice.Message.Content = ""
			choice.FinishReason = geminiFinishReason2OpenAI(candidate.FinishReason)
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *ChatResponse, modelName string, isFirstThoughtChunk *bool, isInThought *bool) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Role = "assistant"

	// Check if this is thought content
	if len(geminiResponse.Candidates) > 0 && len(geminiResponse.Candidates[0].Content.Parts) > 0 {
		parts := geminiResponse.Candidates[0].Content.Parts

		// Tool call chunk
		candidate := geminiResponse.Candidates[0]
		if len(getToolCalls(&candidate)) > 0 {
			choice.Delta.Content = nil
			choice.Delta.ToolCalls = getToolCalls(&candidate)
			finish := "tool_calls"
			choice.FinishReason = &finish
			var response openai.ChatCompletionsStreamResponse
			response.Id = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
			response.Created = helper.GetTimestamp()
			response.Object = "chat.completion.chunk"
			response.Model = modelName
			response.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
			return &response
		}

		part := parts[0]
		if part.Thought != nil && *part.Thought {
			// This is thought content
			choice.Delta.ReasoningContent = part.Text
			choice.Delta.Content = ""
		} else {
			choice.Delta.Content = part.Text
		}
	} else {
		choice.Delta.Content = geminiResponse.GetResponseText()
	}

	var response openai.ChatCompletionsStreamResponse
	response.Id = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
	response.Created = helper.GetTimestamp()
	response.Object = "chat.completion.chunk"
	response.Model = modelName
	response.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func embeddingResponseGemini2OpenAI(response *EmbeddingResponse) *openai.EmbeddingResponse {
	openAIEmbeddingResponse := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingResponseItem, 0, len(response.Embeddings)),
		Model:  "gemini-embedding",
		Usage:  model.Usage{TotalTokens: 0},
	}
	for _, item := range response.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     0,
			Embedding: item.Values,
		})
	}
	return &openAIEmbeddingResponse
}

func StreamHandler(c *gin.Context, resp *http.Response, modelName string, promptTokens int) (*model.ErrorWithStatusCode, *model.Usage) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)

	// Track thought state across chunks
	isFirstThoughtChunk := false
	isInThought := false
	
	// Track usage metadata from the last chunk (Gemini only sends complete usageMetadata in the final chunk)
	var lastUsageMetadata *UsageMetadata

	for scanner.Scan() {
		data := scanner.Text()
		data = strings.TrimSpace(data)
		if !strings.HasPrefix(data, "data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\"")

		var geminiResponse ChatResponse
		err := json.Unmarshal([]byte(data), &geminiResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		// Debug: log raw response to see Gemini's format
		if config.DebugEnabled {
			logger.SysLog("Gemini raw response: " + data)
		}
		
		// Capture usageMetadata if present (usually in the last chunk)
		if geminiResponse.UsageMetadata != nil {
			lastUsageMetadata = geminiResponse.UsageMetadata
		}

		response := streamResponseGeminiChat2OpenAI(&geminiResponse, modelName, &isFirstThoughtChunk, &isInThought)
		if response == nil {
			continue
		}

		responseText += response.Choices[0].Delta.StringContent()

		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// Calculate usage: prefer Gemini's native token count, fallback to manual calculation
	var usage model.Usage
	if lastUsageMetadata != nil {
		usage = model.Usage{
			PromptTokens:     lastUsageMetadata.PromptTokenCount,
			CompletionTokens: lastUsageMetadata.CandidatesTokenCount,
			TotalTokens:      lastUsageMetadata.TotalTokenCount,
		}
	} else {
		// Fallback to manual calculation
		completionTokens := openai.CountTokenText(responseText, modelName)
		usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}

	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse ChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	fullTextResponse.Model = modelName
	
	// Use Gemini's native token count if available, otherwise fallback to manual calculation
	var usage model.Usage
	if geminiResponse.UsageMetadata != nil {
		usage = model.Usage{
			PromptTokens:     geminiResponse.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResponse.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResponse.UsageMetadata.TotalTokenCount,
		}
	} else {
		// Fallback to manual calculation for older API versions
		completionTokens := openai.CountTokenText(geminiResponse.GetResponseText(), modelName)
		usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var geminiEmbeddingResponse EmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &geminiEmbeddingResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if geminiEmbeddingResponse.Error != nil {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: geminiEmbeddingResponse.Error.Message,
				Type:    "gemini_error",
				Param:   "",
				Code:    geminiEmbeddingResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseGemini2OpenAI(&geminiEmbeddingResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
