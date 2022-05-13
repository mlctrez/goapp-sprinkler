package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/server"
	"github.com/mlctrez/goapp-sprinkler/ui"
	"github.com/mlctrez/servicego"
)

type twoFactor struct {
	servicego.Defaults
	serverShutdown func(ctx context.Context) error
}

func main() {

	ui.AddRoutes()

	if app.IsClient {
		app.RunWhenOnBrowser()
	} else {
		servicego.Run(&twoFactor{})
	}

}

func (t *twoFactor) Start(_ service.Service) (err error) {
	t.serverShutdown, err = server.Run()
	return err
}

func (t *twoFactor) Stop(_ service.Service) (err error) {
	if t.serverShutdown != nil {

		stopContext, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		err = t.serverShutdown(stopContext)
		if errors.Is(err, context.Canceled) {
			os.Exit(-1)
		}
	}
	return err
}
