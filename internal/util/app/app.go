package app

import (
	"sync"

	"github.com/pocketbase/pocketbase"
)

var app *pocketbase.PocketBase
var appOnce sync.Once

func Get() *pocketbase.PocketBase {
	appOnce.Do(func() {
		app = pocketbase.New()

	})
	return app
}
