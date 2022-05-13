package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	fetch "github.com/mlctrez/wasm-fetch"
)

type Body struct {
	app.Compo
	pins []Pin
}

func (b *Body) OnAppUpdate(ctx app.Context) {
	if ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}

func (b *Body) OnMount(ctx app.Context) {
	ctx.Async(b.refreshPins)
	ctx.Handle("togglePin", b.togglePin)
}

func (b *Body) togglePin(ctx app.Context, action app.Action) {

	pin, err := strconv.Atoi(action.Tags.Get("pin"))
	if err != nil {
		app.Log(err)
		return
	}
	p := b.pins[pin]
	url := fmt.Sprintf("%s/%d/%s", apiUrl(), p.Number, p.OtherState())
	var response *fetch.Response
	response, err = fetch.Fetch(url, &fetch.Opts{Method: http.MethodPost})
	if err != nil {
		app.Log(err)
		return
	}
	m := map[string][]Pin{}
	err = json.Unmarshal(response.Body, &m)
	if err != nil {
		app.Log("unmarshal error ", err)
	}
	b.pins = m["pins"]
}

func (b *Body) Render() app.UI {
	if !b.Mounted() || b.pins == nil {
		return app.Div()
	}
	return app.Div().Body(
		app.Range(b.pins).Slice(func(i int) app.UI { return b.pins[i].UI(b) }),
	)
}

type Pin struct {
	Number int
	Path   string
	State  string
}

func (p Pin) OtherState() string {
	if p.State == "1" {
		return "off"
	}
	return "on"
}

func (p Pin) UI(b *Body) app.UI {
	toggle := func(ctx app.Context, e app.Event) { ctx.NewAction("togglePin", app.T("pin", p.Number)) }
	stateString := "Off"
	if p.State == "1" {
		stateString = "On"
	}
	message := fmt.Sprintf("Pin %d %s", p.Number, stateString)
	text := app.Span().Style("font-size", "48px").
		Style("padding", "0 25px 0 25px").Text(message)
	button := app.Button().Body(text).OnClick(toggle).
		Style("border-radius", "15px").Style("border-width", "0")

	if stateString == "On" {
		button.Style("background-color", "blue")
	} else {
		button.Style("background-color", "lightgrey")
	}

	return app.Div().Body(
		app.Br(),
		button,
	)
}

func (b *Body) refreshPins() {
	response, err := fetch.Fetch(apiUrl(), &fetch.Opts{
		Method: "GET",
	})
	if err != nil {
		app.Log("fetch error ", err)
	}

	m := map[string][]Pin{}
	err = json.Unmarshal(response.Body, &m)
	if err != nil {
		app.Log("unmarshal error ", err)
	}
	b.pins = m["pins"]
	b.Update()
}

func apiUrl() string {
	url := app.Window().URL()
	return url.Scheme + "://" + url.Host + "/api/pins"
}
