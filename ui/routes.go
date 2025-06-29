package ui

import "github.com/maxence-charriere/go-app/v10/pkg/app"

func AddRoutes() {
	app.Route("/", func() app.Composer {
		return &Body{}
	})
}
