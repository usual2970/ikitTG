package routes

import (
	botUC "ikit-api/internal/usecase/bot"

	"github.com/pocketbase/pocketbase/core"
)

func OnessayUpdate(e *core.RecordUpdateEvent) error {

	uc := botUC.NewessayUsecase()

	return uc.CreateTelegraph(e.HttpContext.Request().Context(), e.Record.Id)
}

func OnessayCreate(e *core.RecordCreateEvent) error {

	uc := botUC.NewessayUsecase()

	return uc.CreateTelegraph(e.HttpContext.Request().Context(), e.Record.Id)
}
