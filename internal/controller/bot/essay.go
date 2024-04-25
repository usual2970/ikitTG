package bot

import (
	"ikit-api/internal/domain"

	"github.com/labstack/echo/v5"
)

type essayController struct {
	uc domain.IessayUsecase
}

func (c *essayController) Notify(ctx echo.Context) error {
	req := &domain.TtsAsyncResp{}
	if err := ctx.Bind(req); err != nil {
		return err
	}
	return c.uc.Notify(ctx.Request().Context(), req)
}
