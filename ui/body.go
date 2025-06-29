package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	fetch "github.com/mlctrez/wasm-fetch"
)

type Body struct {
	app.Compo
	pins  []Pin
	onOff Pin
}

func (b *Body) OnAppUpdate(ctx app.Context) {
	if ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}

func (b *Body) OnMount(ctx app.Context) {
	//ctx.Async(b.refreshPins)
	b.refreshPins()
	ctx.Handle("togglePin", b.togglePin)
}

func (b *Body) togglePin(ctx app.Context, action app.Action) {

	pin, err := strconv.Atoi(action.Tags.Get("pin"))
	if err != nil {
		app.Log(err)
		return
	}
	var p Pin
	var url string
	fmt.Println("Pin is", pin)
	if pin == -1 {
		p = b.pins[6]
		url = fmt.Sprintf("%s/schedule/%s", apiUrl(), p.OtherState())
	} else {
		p = b.pins[pin]
		url = fmt.Sprintf("%s/pins/%d/%s", apiUrl(), p.Number, p.OtherState())
	}
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
	div := app.Div().Style("text-align", "center")
	if !b.Mounted() || b.pins == nil {
		return div.Body(app.H2().Text("No Data. Is Wifi On?"))
	}

	var buttons []app.UI

	for _, pin := range b.pins {
		buttons = append(buttons, pin.UI())
	}

	return div.Body(buttons...)
}

type Pin struct {
	Number int
	Path   string
	State  string
}

func (p Pin) OtherState() string {
	if p.Number == -1 {
		if p.State == "running" {
			return "off"
		}
		return "on"
	}
	if p.State == "1" {
		return "off"
	}
	return "on"
}

func (p Pin) UI() app.UI {

	toggle := func(ctx app.Context, e app.Event) { ctx.NewAction("togglePin", app.T("pin", p.Number)) }
	stateString := "Off"
	if p.State == "1" {
		stateString = "On"
	}
	message := fmt.Sprintf("Pin %d %s", p.Number, stateString)

	if p.Number == -1 {
		message = p.State
	}

	text := app.Span().Style("font-size", "48px").
		Style("padding", "0 25px 0 25px").Text(message)
	button := app.Button().Body(text).OnClick(toggle).
		Style("border-radius", "15px").Style("border-width", "0")

	if stateString == "On" || stateString == "running" {
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
	response, err := fetch.Fetch(apiUrl()+"/pins", &fetch.Opts{Method: "GET"})
	if err != nil {
		app.Log("fetch error ", err)
		return
	}

	m := map[string][]Pin{}
	err = json.Unmarshal(response.Body, &m)
	if err != nil {
		app.Log("unmarshal error ", err)
		return
	}
	b.pins = m["pins"]
	//b.Update()
}

func apiUrl() string {
	url := app.Window().URL()
	return url.Scheme + "://" + url.Host + "/api"
}
