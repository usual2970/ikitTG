package zhipu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ikit-api/internal/util/app"
	xhttp "ikit-api/internal/util/http"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	embeddingUrl   = "https://open.bigmodel.cn/api/paas/v4/embeddings"
	completionUrl  = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	generateImgUrl = "https://open.bigmodel.cn/api/paas/v4/images/generations"
)

const (
	roleTypeUser schema.ChatMessageType = "user"
)

const defaultCompletionModel = "GLM-4"

type completionReq struct {
	Model       string    `json:"model,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	RequestId   string    `json:"request_id,omitempty"`
	DoSample    bool      `json:"do_sample,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_tool,omitempty"`
}

type completionResp struct {
	Created   int       `json:"created"`
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	RequestID string    `json:"request_id"`
	Choices   []Choices `json:"choices"`
	Usage     Usage     `json:"usage"`
}

type embeddingReq struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResp struct {
	Model  string `json:"model"`
	Data   []Data `json:"data"`
	Object string `json:"object"`
	Usage  Usage  `json:"usage"`
}
type Data struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

type Choices struct {
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
	Message      Message `json:"message"`
}
type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Message struct {
	Role       schema.ChatMessageType `json:"role"`
	Content    string                 `json:"content"`
	ToolCallId string                 `json:"tool_call_id"`
	ToolCalls  []ToolCall             `json:"tool_calls"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Index    int          `json:"index"`
	Type     string       `json:"type"`
	Function FunctionResp `json:"function"`
}
type FunctionResp struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

var cache *expirable.LRU[string, string]

var cacheOnce sync.Once

func newZhipuTokenCache() *expirable.LRU[string, string] {

	cacheOnce.Do(func() {
		cache = expirable.NewLRU[string, string](5, nil, time.Hour*defaultHour)
	})

	return cache
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type Zhipu struct {
	apiKey         string
	defaultOptions []llms.CallOption
}

func NewZhipu(apiKey string, options ...llms.CallOption) *Zhipu {

	return &Zhipu{
		defaultOptions: options,
		apiKey:         apiKey,
	}
}

func (z *Zhipu) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	generations, err := z.GenerateContent(ctx, []llms.MessageContent{llms.TextParts(roleTypeUser, prompt)}, options...)
	if err != nil {
		return "", err
	}
	return generations.Choices[0].Content, nil
}

func (z *Zhipu) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	msg, _ := jsoniter.MarshalToString(messages)
	println(msg)
	option := &llms.CallOptions{}
	for _, opt := range z.defaultOptions {
		opt(option)
	}
	for _, opt := range options {
		opt(option)
	}

	if option.Model == "" {
		option.Model = defaultCompletionModel
	}
	token, err := GenerateToken(z.apiKey, time.Hour*12)
	if err != nil {
		return nil, err
	}

	header := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	stopWrods := []string{}
	if len(option.StopWords) > 0 {
		stopWrods = []string{option.StopWords[0]}
	}

	req := &completionReq{
		Model:       option.Model,
		Messages:    []Message{},
		RequestId:   uuid.New().String(),
		Temperature: option.Temperature,
		TopP:        option.TopP,
		Stop:        stopWrods,
	}
	for _, prompt := range messages {
		content := ""
		for _, part := range prompt.Parts {
			switch t := part.(type) {
			case llms.TextContent:
				content += t.Text
			}
		}

		defaultRoleType := roleTypeUser
		switch prompt.Role {

		case schema.ChatMessageTypeAI:
			defaultRoleType = "assistant"
		case schema.ChatMessageTypeSystem:
			defaultRoleType = "system"
		case schema.ChatMessageTypeHuman:
			defaultRoleType = "user"
		case schema.ChatMessageTypeFunction:
			defaultRoleType = "user"
		}

		req.Messages = append(req.Messages, Message{
			Role:    defaultRoleType,
			Content: content,
		})
	}

	if len(option.Functions) > 0 {
		for _, function := range option.Functions {
			req.Tools = append(req.Tools, Tool{
				Type: "function",
				Function: Function{
					Name:        function.Name,
					Description: function.Description,
					Parameters:  function.Parameters,
				},
			})
		}

		req.ToolChoice = "auto"
	}

	bts, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := xhttp.Req(completionUrl, http.MethodPost, bytes.NewReader(bts), header)
	app.Get().Logger().Info("zhipu resp", "req", string(bts), "resp", string(resp), "err", err)
	if err != nil {
		return nil, err
	}

	temp := &completionResp{}

	if err := json.Unmarshal(resp, temp); err != nil {
		return nil, err
	}
	choices := make([]*llms.ContentChoice, 0)
	for _, choice := range temp.Choices {
		if len(choice.Message.ToolCalls) > 0 {
			choices = append(choices, &llms.ContentChoice{
				FuncCall: &schema.FunctionCall{
					Name:      choice.Message.ToolCalls[0].Function.Name,
					Arguments: choice.Message.ToolCalls[0].Function.Arguments,
				},
				StopReason: choice.FinishReason,
				Content:    choice.Message.Content,
			})
			continue
		}

		choices = append(choices, &llms.ContentChoice{
			StopReason: choice.FinishReason,
			Content:    choice.Message.Content,
		})
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil

}

func (z *Zhipu) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	token, err := GenerateToken(z.apiKey, time.Hour*12)
	if err != nil {
		return nil, err
	}

	header := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	rs := make([][]float32, 0)
	for _, text := range texts {
		req := &embeddingReq{
			Input: text,
			Model: "embedding-2",
		}
		bts, _ := json.Marshal(req)
		resp, err := xhttp.Req(embeddingUrl, http.MethodPost, bytes.NewReader(bts), header)
		if err != nil {
			return nil, err
		}

		temp := &embeddingResp{}
		if err := json.Unmarshal(resp, temp); err != nil {
			return nil, err
		}

		for _, data := range temp.Data {
			rs = append(rs, data.Embedding)
		}

	}
	return rs, nil
}

type GenerateImgReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type GenerateImgResp struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func (z *Zhipu) GenerateImg(ctx context.Context, prompt string) (string, error) {
	token, err := GenerateToken(z.apiKey, time.Hour*12)
	if err != nil {
		return "", err
	}

	header := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	req := &GenerateImgReq{
		Model:  "cogview-3",
		Prompt: prompt,
	}

	bts, _ := json.Marshal(req)
	resp, err := xhttp.Req(generateImgUrl, http.MethodPost, bytes.NewReader(bts), header)
	if err != nil {
		return "", err
	}

	temp := &GenerateImgResp{}
	if err := json.Unmarshal(resp, temp); err != nil {
		return "", err
	}

	if len(temp.Data) == 0 {
		return "", errors.New("生成图片失败")
	}

	return temp.Data[0].URL, nil
}

const (
	defaultHour = 12
)

type ZpClaims struct {
	APIKey    string `json:"api_key"`
	Exp       int64  `json:"exp"`
	Timestamp int64  `json:"timestamp"`
}

const tokenKey = "zpToken"

func GenerateToken(apiKey string, duration time.Duration) (string, error) {

	cache := newZhipuTokenCache()

	rs, ok := cache.Get(tokenKey)
	if ok {
		return rs, nil
	}

	if apiKey == "" {
		return "", errors.New("密钥不能为空")
	}
	if !strings.Contains(apiKey, ".") {
		return "", errors.New("密钥格式不正确")
	}

	apiKeyInfo := strings.Split(apiKey, ".")
	key, secret := apiKeyInfo[0], apiKeyInfo[1]

	if duration == 0 {
		duration = defaultHour * time.Hour
	}

	token, err := createToken(ZpClaims{
		key,
		time.Now().Add(duration).Unix(),
		time.Now().Unix(),
	}, secret)
	if err != nil {
		return "", err
	}

	cache.Add(tokenKey, token)
	return token, nil
}

// createToken 生成一个token.
func createToken(claims ZpClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"api_key":   claims.APIKey,
		"exp":       claims.Exp,
		"timestamp": claims.Timestamp,
	})

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	res, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return res, nil
}
