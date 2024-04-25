package bot

import (
	"ikit-api/internal/domain"
	"ikit-api/internal/util/resp"

	"github.com/labstack/echo/v5"
)

type controller struct {
	uc domain.IBotUsecase
}

func (c *controller) Start(ctx echo.Context) error {
	c.uc.Start(ctx.Request().Context())
	return resp.Succ(ctx, nil)
}

func (c *controller) Stop(ctx echo.Context) error {
	c.uc.Stop(ctx.Request().Context())
	return resp.Succ(ctx, nil)
}

func Register(route *echo.Echo, uc domain.IBotUsecase, essayUc domain.IessayUsecase) {
	c := &controller{uc: uc}

	group := route.Group("/api/v1/bot")
	group.POST("/start", c.Start)
	group.POST("/stop", c.Stop)

	essayController := &essayController{uc: essayUc}
	group.POST("/notify", essayController.Notify)

}
