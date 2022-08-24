package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/beagleio"
	"github.com/mlctrez/goapp-sprinkler/schedule"
	"github.com/mlctrez/goapp-sprinkler/server"
	"github.com/mlctrez/goapp-sprinkler/ui"
	"github.com/mlctrez/servicego"
)

type svc struct {
	servicego.Defaults
	serverShutdown func(ctx context.Context) error
	api            *beagleio.Api
	schedule       *schedule.Schedule
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

	defer func() {
		if err != nil {
			t.Log().Error(err)
		}
	}()
	t.api = beagleio.New()

	t.schedule, err = schedule.New()
	if err != nil {
		return err
	}

	t.serverShutdown, err = server.Run(t.schedule)
	return err
}

func (t *svc) Stop(_ service.Service) (err error) {

	if t.serverShutdown != nil {

		stopContext, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		err = t.serverShutdown(stopContext)
		if errors.Is(err, context.Canceled) {
			os.Exit(-1)
		}
	}

	if t.schedule != nil {
		t.schedule.Stop()
	}

	if t.api != nil {
		t.api.Shutdown()
	}

	return err
}
