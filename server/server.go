//go:build !wasm

package server

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/beagleio"
)

//go:embed web/*
var webDirectory embed.FS

func Run() (shutdownFunc func(ctx context.Context) error, err error) {

	address := os.Getenv("ADDRESS")
	if address == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000"
		}
		address = "localhost:" + port
	}

	var listener net.Listener
	if listener, err = net.Listen("tcp4", address); err != nil {
		return nil, err
	}

	if IsDevelopment() {
		fmt.Printf("running on http://%s\n", address)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(gin.Logger(), gin.Recovery(), brotli.Brotli(brotli.DefaultCompression))

	api := beagleio.New()
	if err = api.InitializePins(); err != nil {
		return nil, err
	}
	httpApi := &beagleio.HttpApi{Api: api}
	httpApi.Routes(engine)

	staticHandler := http.FileServer(http.FS(webDirectory))
	engine.GET("/web/:path", gin.WrapH(staticHandler))

	engine.NoRoute(gin.WrapH(BuildHandler()))
	engine.RedirectTrailingSlash = false

	server := &http.Server{Handler: engine}

	go func() {
		var serveErr error
		if strings.HasSuffix(listener.Addr().String(), ":443") {
			serveErr = server.ServeTLS(listener, "cert.pem", "cert.key")
		} else {
			serveErr = server.Serve(listener)
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	return server.Shutdown, nil
}

func BuildHandler() *app.Handler {
	return &app.Handler{
		Author:      "TODO",
		Description: "go-app starter",
		Name:        "go-app starter",
		Scripts:     []string{},
		Icon: app.Icon{
			AppleTouch: "/web/logo-192.png",
			Default:    "/web/logo-192.png",
			Large:      "/web/logo-512.png",
		},
		AutoUpdateInterval: autoUpdateInterval(),
		ShortName:          "starter",
		Version:            getRuntimeVersion(),
		Styles:             []string{},
		Title:              "go-app starter",
	}
}
