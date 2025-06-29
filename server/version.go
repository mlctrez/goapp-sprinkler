package server

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

var (
	Version        string
	Commit         string
	runtimeVersion string
)

const GoAppVersionKey = "GOAPP_VERSION"

func GoAppVersion() string {
	return app.Getenv(GoAppVersionKey)
}

func getRuntimeVersion() string {
	if runtimeVersion == "" {
		if IsDevelopment() {
			t := time.Now().UTC().String()
			runtimeVersion = fmt.Sprintf(`%x`, sha1.Sum([]byte(t)))
		} else {
			runtimeVersion = fmt.Sprintf("%s@%s", Version, Commit)
		}
	}
	return runtimeVersion
}

func IsDevelopment() bool {
	return os.Getenv("DEV") != ""
}

func autoUpdateInterval() time.Duration {
	if IsDevelopment() {
		return 3 * time.Second
	}
	return 24 * time.Hour
}
