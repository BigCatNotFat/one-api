package model

type Tool struct {
	Id           string                 `json:"id,omitempty"`
	Type         string                 `json:"type,omitempty"` // when splicing claude tools stream messages, it is empty
	Function     Function               `json:"function"`
	ExtraContent map[string]interface{} `json:"extra_content,omitempty"` // For Gemini 3 thought_signature
}

type Function struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`       // when splicing claude tools stream messages, it is empty
	Parameters  any    `json:"parameters,omitempty"` // request
	Arguments   any    `json:"arguments,omitempty"`  // response
}
