package vertexai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor/gemini"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/relaymode"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var ModelList = []string{
	"gemini-pro", "gemini-pro-vision",
	"gemini-exp-1206",
	"gemini-1.5-pro-001", "gemini-1.5-pro-002",
	"gemini-1.5-flash-001", "gemini-1.5-flash-002",
	"gemini-2.0-flash-exp", "gemini-2.0-flash-001",
	"gemini-2.0-flash-lite-preview-02-05",
	"gemini-2.0-flash-thinking-exp-01-21",
	// Gemini 2.5
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	// Gemini 3 (preview)
	"gemini-3-flash-preview",
	"gemini-3-pro-preview",
	"gemini-3-pro-image-preview",
}

type Adaptor struct {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	geminiRequest, err := gemini.ConvertRequest(*request)
	if err != nil {
		return nil, err
	}
	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, geminiRequest)
	return geminiRequest, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText = gemini.StreamHandler(c, resp, meta.ActualModelName)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = gemini.EmbeddingHandler(c, resp)
		default:
			err, usage = gemini.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}
