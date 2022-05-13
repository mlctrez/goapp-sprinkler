package beagleio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Api struct {
	GpioPaths []string
}

func New() *Api {

	api := &Api{
		GpioPaths: []string{
			// 6 0 P8.8 /sys/class/gpio/gpio67
			"/sys/class/gpio/gpio67",
			// 8 1 P8.10 /sys/class/gpio/gpio68
			"/sys/class/gpio/gpio68",
			// 10 2 P8.12 /sys/class/gpio/gpio44
			"/sys/class/gpio/gpio44",
			// 12 3 P8.14 /sys/class/gpio/gpio26
			"/sys/class/gpio/gpio26",
			// 14 4 P8.16 /sys/class/gpio/gpio46
			"/sys/class/gpio/gpio46",
			// 16 5 P8.18 /sys/class/gpio/gpio65
			"/sys/class/gpio/gpio65",
		},
	}

	if os.Getenv("DEV") != "" {
		var paths []string
		for i := 0; i < 6; i++ {
			p := fmt.Sprintf("temp/gpio/pin%d", i)
			_ = os.MkdirAll(p, 0755)
			_ = os.WriteFile(filepath.Join(p, "direction"), []byte("unset"), 0666)
			_ = os.WriteFile(filepath.Join(p, "value"), []byte("unset"), 0666)
			paths = append(paths, p)
		}
		api = &Api{GpioPaths: paths}
	}

	return api
}

func writeString(path, value string) {

	fmt.Println("writeString", path, value)

	var f *os.File
	var err error

	mode := os.O_WRONLY | os.O_TRUNC

	if f, err = os.OpenFile(path, mode, 0666); err != nil {
		log.Println("writeString os.OpenFile", err)
		return
	}

	defer func() {
		errClose := f.Close()
		if errClose != nil {
			log.Println("writeString f.Close()", errClose)
		}
	}()

	_, err = f.Write([]byte(value))
	if err != nil {
		log.Println("writeString f.Write", err)
	}

	return
}

func (a *Api) InitializePins() error {

	for _, path := range a.GpioPaths {
		writeString(filepath.Join(path, "direction"), "out")
	}

	a.PinsOff()

	return nil
}

func (a *Api) ChangePin(pin, state string) {
	thePin, err := strconv.Atoi(pin)
	if err != nil {
		return
	}
	if thePin >= 0 && thePin < len(a.GpioPaths) {
		path := filepath.Join(a.GpioPaths[thePin], "value")
		switch strings.ToLower(state) {
		case "on", "true":
			writeString(path, "1")
		default:
			writeString(path, "0")
		}
	}
}

func (a *Api) ReadPin(thePin int) (state string) {
	var err error
	if thePin >= 0 && thePin < len(a.GpioPaths) {
		path := filepath.Join(a.GpioPaths[thePin], "value")
		var data []byte
		if data, err = os.ReadFile(path); err == nil {
			return string(data)
		} else {
			fmt.Println("ReadPin error", err)
		}
	}
	return ""
}

func (a *Api) PinsOff() {
	for _, p := range a.GpioPaths {
		writeString(filepath.Join(p, "value"), "0")
	}
}

func (a *Api) Shutdown() {
	a.PinsOff()
}
