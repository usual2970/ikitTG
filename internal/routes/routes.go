package routes

import (
	botUC "ikit-api/internal/usecase/bot"

	"github.com/labstack/echo/v5"

	"ikit-api/internal/controller/bot"
)

func Route(echo *echo.Echo) {

	uc, err := botUC.New()
	if err != nil {
		panic(err)
	}

	essayUc := botUC.NewessayUsecase()

	bot.Register(echo, uc, essayUc)
}
