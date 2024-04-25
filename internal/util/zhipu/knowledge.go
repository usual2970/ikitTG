package zhipu

import (
	"bytes"
	"fmt"
	xhttp "ikit-api/internal/util/http"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type knowledge struct {
	apiKey string
}

func NewKnowledge(apiKey string) *knowledge {
	return &knowledge{
		apiKey: apiKey,
	}
}

type KnowledgeInvokeReq struct {
	RequestId    string       `json:"request_id,omitempty"`
	Prompt       []PromptItem `json:"prompt"`
	ReturnType   string       `json:"return_type,omitempty"`
	KnowledgeIds []int        `json:"knowledge_ids,omitempty"`
	DocumentIds  []int        `json:"document_ids,omitempty"`
}

type KnowledgeInvokeResp struct {
	Data struct {
		RequestId string `json:"requestId"`
		Content   string `json:"content"`
	} `json:"data"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type PromptItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const baseUrl = "https://open.bigmodel.cn/api/llm-application/open"

func (k *knowledge) Invoke(content string) (*KnowledgeInvokeResp, error) {
	req := &KnowledgeInvokeReq{

		Prompt: []PromptItem{
			{
				Role:    "user",
				Content: content,
			},
		},
	}

	url := fmt.Sprintf("%s/model-api/%s/invoke", baseUrl, "1764820437164064769")

	reqBytes, _ := jsoniter.Marshal(req)

	token, err := GenerateToken(k.apiKey, time.Hour*12)
	if err != nil {
		return nil, err
	}

	header := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	resp, err := xhttp.Req(url, http.MethodPost, bytes.NewBuffer(reqBytes), header)
	if err != nil {
		return nil, err
	}

	rs := &KnowledgeInvokeResp{}
	if err := jsoniter.Unmarshal(resp, rs); err != nil {
		return nil, err
	}

	if rs.Code != 200 {
		return nil, fmt.Errorf("code:%d,msg:%s", rs.Code, rs.Message)
	}
	return rs, nil

}
