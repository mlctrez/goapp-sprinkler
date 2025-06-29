//go:build !wasm

package server

import (
	"context"
	"embed"
	"fmt"
	brotli "github.com/anargu/gin-brotli"
	"github.com/mlctrez/goapp-sprinkler/schedule"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/mlctrez/goapp-sprinkler/beagleio"
)

//go:embed web/*
var webDirectory embed.FS

func Run(s *schedule.Schedule) (shutdownFunc func(ctx context.Context) error, err error) {

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

	compresser := brotli.Brotli(brotli.DefaultCompression)
	_ = compresser
	if IsDevelopment() {
		engine.Use(gin.Logger(), gin.Recovery())
	} else {
		engine.Use(gin.Recovery())
	}

	api := beagleio.New()
	if err = api.InitializePins(); err != nil {
		return nil, err
	}
	httpApi := &beagleio.HttpApi{Api: api, StartStop: s.StartStop}
	httpApi.Routes(engine)

	staticHandler := http.FileServer(http.FS(webDirectory))
	engine.GET("/web/:path", gin.WrapH(staticHandler))

	handler := BuildHandler()
	ginHandler := gin.WrapH(handler)
	engine.GET("/", ginHandler)
	engine.GET("/app.css", ginHandler)
	engine.GET("/app.js", ginHandler)
	engine.GET("/app-worker.js", ginHandler)
	engine.GET("/wasm_exec.js", ginHandler)
	engine.GET("/manifest.webmanifest", ginHandler)

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
		Description: "sprinkler",
		Name:        "sprinkler",
		Scripts:     []string{},
		Icon: app.Icon{
			Default: "/web/logo-192.png",
			Large:   "/web/logo-512.png",
			SVG:     "/web/logo-512.svg",
		},
		ShortName:       "sprinkler",
		Version:         getRuntimeVersion(),
		Styles:          []string{},
		Title:           "sprinkler",
		BackgroundColor: "#222",
	}
}
