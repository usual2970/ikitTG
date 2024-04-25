package main

import (
	"ikit-api/internal/routes"
	"ikit-api/internal/util/app"
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "ikit-api/migrations"

	_ "github.com/pocketbase/pocketbase/migrations"
)

func main() {
	app := app.Get()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	app.OnRecordAfterCreateRequest("essay").Add(func(e *core.RecordCreateEvent) error {

		return routes.OnessayCreate(e)
	})

	app.OnRecordAfterUpdateRequest("essay").Add(func(e *core.RecordUpdateEvent) error {

		return routes.OnessayUpdate(e)
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		routes.Route(e.Router)

		if err := routes.Register(); err != nil {
			return err
		}
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
