package gemini

type ChatRequest struct {
	Contents          []ChatContent        `json:"contents"`
	SafetySettings    []ChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig  ChatGenerationConfig `json:"generation_config,omitempty"`
	Tools             []ChatTools          `json:"tools,omitempty"`
	ToolConfig        *ToolConfig          `json:"tool_config,omitempty"`
	SystemInstruction *ChatContent         `json:"system_instruction,omitempty"`
}

type EmbeddingRequest struct {
	Model                string      `json:"model"`
	Content              ChatContent `json:"content"`
	TaskType             string      `json:"taskType,omitempty"`
	Title                string      `json:"title,omitempty"`
	OutputDimensionality int         `json:"outputDimensionality,omitempty"`
}

type BatchEmbeddingRequest struct {
	Requests []EmbeddingRequest `json:"requests"`
}

type EmbeddingData struct {
	Values []float64 `json:"values"`
}

type EmbeddingResponse struct {
	Embeddings []EmbeddingData `json:"embeddings"`
	Error      *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type FunctionCall struct {
	FunctionName string `json:"name"`
	Arguments    any    `json:"args"`
	// Note: thoughtSignature should NOT be inside functionCall, it belongs to Part level
}

type FunctionResponse struct {
	FunctionName string `json:"name"`
	Response     any    `json:"response"`
}

type Part struct {
	Text             string            `json:"text,omitempty"`
	InlineData       *InlineData       `json:"inlineData,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResp     *FunctionResponse `json:"functionResponse,omitempty"`
	Thought          *bool             `json:"thought,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"` // Gemini 3: thought signature for function calls
}

type ChatContent struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type ChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type ChatTools struct {
	FunctionDeclarations any `json:"function_declarations,omitempty"`
}

type ChatGenerationConfig struct {
	ResponseMimeType string                 `json:"responseMimeType,omitempty"`
	ResponseSchema   any                    `json:"responseSchema,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"topP,omitempty"`
	TopK             float64                `json:"topK,omitempty"`
	MaxOutputTokens  int                    `json:"maxOutputTokens,omitempty"`
	CandidateCount   int                    `json:"candidateCount,omitempty"`
	StopSequences    []string               `json:"stopSequences,omitempty"`
	ThinkingConfig   map[string]interface{} `json:"thinkingConfig,omitempty"`
}

type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"function_calling_config,omitempty"`
}

type FunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"`                   // AUTO / ANY / NONE
	AllowedFunctionNames []string `json:"allowed_function_names,omitempty"` // optional constraint
}
