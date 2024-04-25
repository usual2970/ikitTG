package domain

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type IBotUsecase interface {
	Process(ctx context.Context) error

	Start(ctx context.Context)

	Stop(ctx context.Context)
}

type Essay struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	FileId    string `json:"fileId"`
	TaskId    string `json:"taskId"`
	File      string `json:"file"`
	Thumb     string `json:"thumb"`
	Telegraph string `json:"telegraph"`
	Meta
}

type ListessayReq struct {
	Offset int
	Limit  int
	Filter string
}

type AddessayReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type TtsAsyncReq struct {
	Playload TtsAsyncPayload `json:"payload"`

	Context TtsAsyncContext `json:"context"`

	Header TtsAsyncHeader `json:"header"`
}

type TtsAsyncContext struct {
	DeviceID string `json:"device_id"`
}

type TtsAsyncHeader struct {
	Appkey string `json:"appkey"`
	Token  string `json:"token"`
}

type TtsAsyncPayload struct {
	TtsRequest   TtsAsyncRequest `json:"tts_request"`
	EnableNotify bool            `json:"enable_notify"`
	NotifyUrl    string          `json:"notify_url"`
}

type TtsAsyncRequest struct {
	Voice         string `json:"voice"`
	SampleRate    int    `json:"sample_rate"`
	Format        string `json:"format"`
	Text          string `json:"text"`
	EnableSubtile bool   `json:"enable_subtitle"`
}

type TtsAsyncResp struct {
	Status int `json:"status"`

	ErrorCode int `json:"error_code"`

	ErrorMessage string `json:"error_message"`

	RequestId string `json:"request_id"`

	Data TtsAsyncRespData `json:"data"`

	Url string `json:"url"`
}

type TtsAsyncRespData struct {
	TaskId       string             `json:"task_id"`
	Sentences    []TtsAsyncSentence `json:"sentences"`
	AudioAddress string             `json:"audio_address"`
	NotifyCustom string             `json:"notify_custom"`
}

type TtsAsyncSentence struct {
	Text      string `json:"text"`
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
}

type TgChatItem struct {
	Chat     tgbotapi.Chattable
	Callback func(message tgbotapi.Message) error
}

type TgCallback func(message tgbotapi.Message) error

func NewTgChatItem(chat tgbotapi.Chattable, callback ...TgCallback) *TgChatItem {
	var cb TgCallback
	if len(callback) > 0 {
		cb = callback[0]
	}
	return &TgChatItem{
		Chat:     chat,
		Callback: cb,
	}
}

type IessayUsecase interface {
	Add(ctx context.Context, req *AddessayReq) error
	List(ctx context.Context, req *ListessayReq) ([]Essay, error)
	Delete(ctx context.Context, id string) error
	Detail(ctx context.Context, id string) (*Essay, error)
	Notify(ctx context.Context, req *TtsAsyncResp) error
	UpdateFileId(ctx context.Context, id string, fileId string) error

	CreateTelegraph(ctx context.Context, id string) error
}
