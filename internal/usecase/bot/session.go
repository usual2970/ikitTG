package bot

import (
	"context"
	"errors"
	"fmt"
	"ikit-api/internal/domain"
	"ikit-api/internal/util/app"
	xhttp "ikit-api/internal/util/http"
	"net/http"
	"regexp"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	KindAddessay = "essay"
)

const (
	StateWaitTitle   = "wait_title"
	SteteWaitContent = "wait_content"
)

const perPageSize = 10

type Session struct {
	ChatID int64
	Kind   string
	State  string // 添加文章

	bot *tgbotapi.BotAPI

	essay *domain.AddessayReq
}

func NewSession(chatID int64, bot *tgbotapi.BotAPI) *Session {
	return &Session{
		ChatID: chatID,
		bot:    bot,
	}
}

func (s *Session) Process(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {

	return s.processUpdate(ctx, update)
}

func (s *Session) processUpdate(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {
	// 处理命令
	if update.CallbackQuery != nil {
		return s.processCallback(ctx, update)
	}

	if update.Message.IsCommand() {
		return s.processCommand(ctx, update)
	}

	return s.processText(ctx, update)
}

func (s *Session) processCommand(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {

	msg := update.Message
	switch update.Message.Command() {
	case "start", "menu":
		// 发送欢迎消息
		reply := tgbotapi.NewMessage(msg.From.ID, "欢迎使用英语文章背诵机器人")
		reply.ReplyMarkup = getKeyBoards()

		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil

	}

	return nil, errors.New("unknown command")
}

var detailReg = regexp.MustCompile(`essay:(.+)$`)
var deleteReg = regexp.MustCompile(`delete:(.+)$`)
var nextReg = regexp.MustCompile(`next:(.*)$`)

func (s *Session) processCallback(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {

	data := update.CallbackData()
	switch data {
	case "add":
		// 开始添加文章
		s.Kind = KindAddessay
		s.State = StateWaitTitle

		reply := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "请输入文章标题")

		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil
	case "list":
		essays, err := s.getessayUc().List(ctx, &domain.ListessayReq{
			Filter: "1=1",
			Limit:  perPageSize,
		})
		if err != nil {
			return nil, err
		}

		reply := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "*文章列表*")
		reply.ParseMode = "MarkdownV2"
		reply.ReplyMarkup = getessayListKeyBoards(essays)
		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil

	case "return2menu":
		reply := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "欢迎使用英语文章背诵机器人")
		reply.ReplyMarkup = getKeyBoards()

		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil

	}

	if matches := nextReg.FindStringSubmatch(data); len(matches) == 2 {
		id := matches[1]

		req := &domain.ListessayReq{
			Limit: perPageSize,
		}
		req.Filter = "1=1"
		if id != "" {
			req.Filter = "id < '" + id + "'"
		}
		essays, err := s.getessayUc().List(ctx, req)
		if err != nil {
			return nil, err
		}

		reply := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "*文章列表*")
		reply.ParseMode = "MarkdownV2"
		reply.ReplyMarkup = getessayListKeyBoards(essays)

		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil
	}

	if matches := detailReg.FindStringSubmatch(data); len(matches) == 2 {
		id := matches[1]
		return s.detail(ctx, id, update)
	}

	if matches := deleteReg.FindStringSubmatch(data); len(matches) == 2 {
		id := matches[1]
		return s.delete(ctx, id, update)
	}

	app.Get().Logger().Info("process callback", "data", update.CallbackData(), "query", *update.CallbackQuery)

	return nil, errors.New("unknown command")
}

func (s *Session) delete(ctx context.Context, id string, update tgbotapi.Update) ([]domain.TgChatItem, error) {
	if err := s.getessayUc().Delete(ctx, id); err != nil {
		return nil, err
	}

	essays, err := s.getessayUc().List(ctx, &domain.ListessayReq{
		Filter: "1=1",
		Limit:  perPageSize,
	})
	if err != nil {
		return nil, err
	}

	reply := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "*文章列表*")
	reply.ParseMode = "MarkdownV2"
	reply.ReplyMarkup = getessayListKeyBoards(essays)
	return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil
}

func (s *Session) detail(ctx context.Context, id string, update tgbotapi.Update) ([]domain.TgChatItem, error) {
	essay, err := s.getessayUc().Detail(ctx, id)
	if err != nil {
		return nil, err
	}

	rs := make([]domain.TgChatItem, 0)
	callbacks := make([]domain.TgCallback, 0)
	var audio *tgbotapi.AudioConfig
	if essay.FileId != "" {
		temp := tgbotapi.NewAudio(update.CallbackQuery.From.ID, tgbotapi.FileID(essay.FileId))
		audio = &temp

	} else if essay.File != "" {
		byts, _ := xhttp.Req(essay.File, http.MethodGet, nil, map[string]string{})
		temp := tgbotapi.NewAudio(update.CallbackQuery.From.ID, tgbotapi.FileBytes{
			Name:  essay.Title,
			Bytes: byts,
		})

		if essay.Thumb != "" {
			thumbBytes, _ := xhttp.Req(essay.Thumb, http.MethodGet, nil, map[string]string{})
			temp.Thumb = tgbotapi.FileBytes{
				Name:  essay.Title,
				Bytes: thumbBytes,
			}
		}

		app.Get().Logger().Info("download file", "url", essay.File)
		callbacks = append(callbacks, func(message tgbotapi.Message) error {

			return s.getessayUc().UpdateFileId(ctx, essay.Id, message.Audio.FileID)
		})
		audio = &temp
	}

	if audio != nil {

		audio.Title = essay.Title
		rs = append(rs, *domain.NewTgChatItem(audio, callbacks...))
	}

	tpl := `
*%s*

%s
		`

	var reply tgbotapi.MessageConfig
	if essay.Telegraph != "" {
		reply = tgbotapi.NewMessage(update.CallbackQuery.From.ID, essay.Telegraph)

	} else {
		reply = tgbotapi.NewMessage(update.CallbackQuery.From.ID, fmt.Sprintf(tpl, essay.Title, essay.Content))

		reply.ParseMode = "Markdown"

	}

	reply.ReplyMarkup = getDetailKeyBoards(*essay)
	rs = append(rs, *domain.NewTgChatItem(reply))

	return rs, nil
}

func (s *Session) processText(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {
	switch s.Kind {
	case KindAddessay:
		return s.processessay(ctx, update)
	}

	return nil, errors.New("unknown command")

}

func (s *Session) processessay(ctx context.Context, update tgbotapi.Update) ([]domain.TgChatItem, error) {

	switch s.State {
	case StateWaitTitle:
		s.essay = &domain.AddessayReq{
			Title: update.Message.Text,
		}
		s.State = SteteWaitContent
		reply := tgbotapi.NewMessage(update.Message.From.ID, "请输入文章内容")
		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil
	case SteteWaitContent:
		s.essay.Content = update.Message.Text
		reply := tgbotapi.NewMessage(update.Message.From.ID, "文章已保存")

		reply.ReplyMarkup = getReturnKeyBoards()

		if err := s.getessayUc().Add(ctx, s.essay); err != nil {
			app.Get().Logger().Info("Addessay error", "error", err)
			return nil, err
		}

		app.Get().Logger().Info("essay", "essay", *s.essay)
		s.clearState()

		return []domain.TgChatItem{*domain.NewTgChatItem(reply)}, nil
	}

	return nil, errors.New("unknown command")
}

func (s *Session) getessayUc() domain.IessayUsecase {
	return &essayUsecase{
		bot: s.bot,
	}
}

func (s *Session) clearState() {
	s.State = ""
	s.Kind = ""
	s.essay = nil
}

var sessionMap *sessionList
var once sync.Once

func GetSessions() *sessionList {
	once.Do(func() {
		sessionMap = NewSessionList()
	})

	return sessionMap
}

func ClearSessions() {
	GetSessions().Clear()
}

func GetSession(update tgbotapi.Update, bot *tgbotapi.BotAPI) *Session {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.From.ID
	} else {
		chatID = update.Message.From.ID
	}
	session, ok := GetSessions().GetSession(chatID)
	if !ok {
		session = &Session{
			ChatID: chatID,
			bot:    bot,
		}
		AddSession(session)
	}

	return session
}

func AddSession(session *Session) {
	GetSessions().AddSession(session)
}

type sessionList struct {
	sessions map[int64]*Session
	sync.RWMutex
}

func NewSessionList() *sessionList {
	return &sessionList{
		sessions: make(map[int64]*Session),
	}
}

func (s *sessionList) AddSession(session *Session) {
	s.Lock()
	defer s.Unlock()
	s.sessions[session.ChatID] = session
}

func (s *sessionList) GetSession(chatID int64) (*Session, bool) {
	s.RLock()
	defer s.RUnlock()
	rs, ok := s.sessions[chatID]

	return rs, ok
}

func (s *sessionList) Clear() {
	s.Lock()
	defer s.Unlock()
	s.sessions = make(map[int64]*Session)
}

func getKeyBoards() tgbotapi.InlineKeyboardMarkup {

	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("添加文章", "add"),
			tgbotapi.NewInlineKeyboardButtonData("文章列表", "list"),
		},
	}...)

}

// 将文章列表组织成keyboards
func getessayListKeyBoards(essays []domain.Essay) tgbotapi.InlineKeyboardMarkup {
	rs := make([][]tgbotapi.InlineKeyboardButton, 0)
	for _, e := range essays {
		rs = append(rs, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(e.Title, "essay:"+e.Id)})
	}

	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("返回到菜单", "return2menu"),
	}

	if len(essays) == perPageSize {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("下一页", "next:"+essays[len(essays)-1].Id))
	}

	rs = append(rs, buttons)
	return tgbotapi.NewInlineKeyboardMarkup(rs...)
}

func getReturnKeyBoards() tgbotapi.InlineKeyboardMarkup {

	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("返回到菜单", "return2menu"),
		},
	}...)

}

func getDetailKeyBoards(essay domain.Essay) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("删除文章", "delete:"+essay.Id),
			tgbotapi.NewInlineKeyboardButtonData("返回到列表", "list"),
			tgbotapi.NewInlineKeyboardButtonData("返回到菜单", "return2menu"),
		},
	}...)
}
