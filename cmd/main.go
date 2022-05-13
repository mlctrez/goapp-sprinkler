package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/beagleio"
	"github.com/mlctrez/goapp-sprinkler/server"
	"github.com/mlctrez/goapp-sprinkler/ui"
	"github.com/mlctrez/servicego"
)

type svc struct {
	servicego.Defaults
	serverShutdown func(ctx context.Context) error
	api            *beagleio.Api
}

func main() {

	ui.AddRoutes()

	if app.IsClient {
		app.RunWhenOnBrowser()
	} else {
		servicego.Run(&svc{})
	}

}

func (t *svc) Start(_ service.Service) (err error) {

	t.api = beagleio.New()

	t.serverShutdown, err = server.Run()
	return err
}

func (t *svc) Stop(_ service.Service) (err error) {

	if t.api != nil {
		t.api.Shutdown()
	}

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
