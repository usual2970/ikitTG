package bot

import (
	"context"
	"ikit-api/internal/domain"
	"ikit-api/internal/util/app"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const processNum = 10

var ucOnce sync.Once

var instance domain.IBotUsecase

type usecase struct {
	ch        tgbotapi.UpdatesChannel
	bot       *tgbotapi.BotAPI
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	sync.RWMutex
}

func New() (domain.IBotUsecase, error) {
	ucOnce.Do(func() {
		instance = &usecase{}
	})

	return instance, nil
}

func (u *usecase) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-u.ch:

			session := GetSession(update, u.bot)
			reply, err := session.Process(ctx, update)
			if err != nil {
				app.Get().Logger().Info("process update error: %v", err)
				continue
			}

			for _, item := range reply {
				rs, err := u.bot.Send(item.Chat)
				if err != nil {
					app.Get().Logger().Info("send item error:", "err", err, "rs", rs, "item", item.Chat)
				} else {
					app.Get().Logger().Info("send item success:", "rs", rs, "item", item.Chat)
				}

				if item.Callback != nil && err == nil {
					if err := item.Callback(rs); err != nil {
						app.Get().Logger().Info("send callback error:", "err", err)
					}
				}
			}

		}
	}

}

func (u *usecase) Start(ctx context.Context) {

	u.RLock()
	if u.isRunning {
		u.RUnlock()
		return
	}
	u.RUnlock()

	tgToken := os.Getenv("TG_TOKEN")

	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		app.Get().Logger().Info("start bot error: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(ctx)

	u.Lock()
	u.isRunning = true
	u.bot = bot
	u.ch = bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	u.cancel = cancel
	u.Unlock()

	u.wg.Add(processNum)
	for i := 0; i < processNum; i++ {
		go func() {
			defer u.wg.Done()
			u.Process(ctx)
		}()
	}

	app.Get().Logger().Info("start bot")
}

func (u *usecase) Stop(ctx context.Context) {

	app.Get().Logger().Info("stop bot")
	u.cancel()
	u.bot.StopReceivingUpdates()

	u.wg.Wait()

	ClearSessions()

	u.Lock()
	u.isRunning = false
	u.Unlock()
}
