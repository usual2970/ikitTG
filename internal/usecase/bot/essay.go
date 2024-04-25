package bot

import (
	"context"
	"ikit-api/internal/domain"
	"ikit-api/internal/util/app"
	"ikit-api/internal/util/audio"
	"ikit-api/internal/util/telegraph"
	"ikit-api/internal/util/zhipu"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type essayUsecase struct {
	bot *tgbotapi.BotAPI
}

func NewessayUsecase(bot ...*tgbotapi.BotAPI) domain.IessayUsecase {
	if len(bot) > 0 {
		return &essayUsecase{bot: bot[0]}
	}
	return &essayUsecase{}
}

func (e *essayUsecase) UpdateFileId(ctx context.Context, id string, fileId string) error {
	record, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app.Get(), record)

	form.LoadData(map[string]any{
		"file_id": fileId,
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}

func (e *essayUsecase) CreateTelegraph(ctx context.Context, id string) error {
	record, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	if record.GetString("content") == "" {
		app.Get().Logger().Info("empty content  no need to upload")
		return nil
	}

	tp := telegraph.New()

	imgUrl := ""
	if record.GetString("thumb") != "" {
		imgUrl = app.Get().Settings().Meta.AppUrl + "/api/files/" + record.BaseFilesPath() + "/" + record.GetString("thumb")
	}

	page, err := tp.CreatePage(record.GetString("title"), record.GetString("content"), imgUrl)
	if err != nil {

		app.Get().Logger().Error("create telegraph error:", err)
		return err
	}

	app.Get().Logger().Info("create telegraph success", "page", page)

	// 先重新获取一下record,后面考虑加锁
	cRecord, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app.Get(), cRecord)

	form.LoadData(map[string]any{
		"telegraph": page.URL,
	})

	if err := form.Submit(); err != nil {
		return err
	}
	app.Get().Logger().Info("create telegraph success", "page", page)
	return nil
}

func (e *essayUsecase) text2Speech(ctx context.Context, id string) error {
	record, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	if record.GetString("content") == "" {
		app.Get().Logger().Info("empty content  no need to upload")
		return nil
	}

	resp, err := audio.Azure(ctx, record.GetString("content"))
	if err != nil {
		return err
	}

	f, err := filesystem.NewFileFromBytes(resp, record.GetString("title"))
	if err != nil {
		return err
	}

	// 先重新获取一下record,后面考虑加锁
	cRecord, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app.Get(), cRecord)
	form.AddFiles("file", f)

	if err := form.Submit(); err != nil {
		return err
	}
	return nil
}

func (e *essayUsecase) Notify(ctx context.Context, req *domain.TtsAsyncResp) error {
	app.Get().Logger().Info("essay notify:", "req", req)

	record, err := app.Get().Dao().FindFirstRecordByData("essay", "task_id", req.Data.TaskId)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app.Get(), record)

	form.LoadData(map[string]any{
		"sentences": req.Data.Sentences,
	})

	f1, err := filesystem.NewFileFromUrl(ctx, req.Data.AudioAddress)
	if err != nil {
		return err
	}
	form.AddFiles("file", f1)

	if err := form.Submit(); err != nil {
		return err
	}
	return nil
}

func (e *essayUsecase) Add(ctx context.Context, req *domain.AddessayReq) error {

	collection, err := app.Get().Dao().FindCollectionByNameOrId("essay")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)

	form := forms.NewRecordUpsert(app.Get(), record)

	form.LoadData(map[string]any{
		"title":   req.Title,
		"content": req.Content,
	})

	apiKey := os.Getenv("ZHIPU_API_KEY")
	zp := zhipu.NewZhipu(apiKey)

	url, err := zp.GenerateImg(ctx, req.Title)
	if err != nil {
		return err
	}
	f, _ := filesystem.NewFileFromUrl(ctx, url)

	form.AddFiles("thumb", f)

	// 保存到数据库
	if err := form.Submit(); err != nil {
		return err
	}

	go func() {
		if err := e.CreateTelegraph(context.Background(), record.Id); err != nil {
			app.Get().Logger().Error("createTelegraph error:", "err", err)
		} else {
			app.Get().Logger().Info("success createTelegraph", "id", record.Id)
		}
	}()

	// 文字转换成语音
	go func() {
		if err := e.text2Speech(context.Background(), record.Id); err != nil {
			app.Get().Logger().Error("text2speech error:", "err", err)
		} else {
			app.Get().Logger().Info("success text2speech", "id", record.Id)
		}
	}()

	return nil
}

func (e *essayUsecase) List(ctx context.Context, req *domain.ListessayReq) ([]domain.Essay, error) {

	records, err := app.Get().Dao().FindRecordsByFilter("essay", req.Filter, "-created", req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	rs := make([]domain.Essay, 0, len(records))
	for _, record := range records {
		rs = append(rs, domain.Essay{
			Meta: domain.Meta{
				Id:      record.Id,
				Created: record.GetTime("created"),
				Updated: record.GetTime("updated"),
			},
			Title: record.GetString("title"),
		})
	}

	return rs, nil
}

func (e *essayUsecase) Delete(ctx context.Context, id string) error {
	record, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return err
	}

	return app.Get().Dao().DeleteRecord(record)
}

func (e *essayUsecase) Detail(ctx context.Context, id string) (*domain.Essay, error) {
	record, err := app.Get().Dao().FindRecordById("essay", id)
	if err != nil {
		return nil, err
	}
	file := ""
	if record.GetString("file") != "" {
		file = app.Get().Settings().Meta.AppUrl + "/api/files/" + record.BaseFilesPath() + "/" + record.GetString("file")
	}

	thumb := ""
	if record.GetString("thumb") != "" {
		thumb = app.Get().Settings().Meta.AppUrl + "/api/files/" + record.BaseFilesPath() + "/" + record.GetString("thumb")
	}
	rs := &domain.Essay{
		Meta: domain.Meta{
			Id:      record.Id,
			Created: record.GetTime("created"),
			Updated: record.GetTime("updated"),
		},
		Title:     record.GetString("title"),
		Content:   record.GetString("content"),
		File:      file,
		FileId:    record.GetString("file_id"),
		Thumb:     thumb,
		Telegraph: record.GetString("telegraph"),
	}
	return rs, nil
}
