package ui

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/server"
)

type Body struct {
	app.Compo
}

func (b *Body) Render() app.UI {
	return app.Div().Text(server.GoAppVersion())
}
