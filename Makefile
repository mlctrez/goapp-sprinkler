
APP_NAME=sprinkler

VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "0.0.0")
COMMIT=$(shell git rev-parse --short HEAD || echo "HEAD")
MODULE=$(shell grep ^module go.mod | awk '{print $$2;}')
LD_FLAGS="-w -X $(MODULE)/server.Version=$(VERSION) -X $(MODULE)/server.Commit=$(COMMIT)"

run: binary
	@DEV=1 NATS_SERVER=nats://goservice:19201 ./temp/$(APP_NAME)

wasm:
	@mkdir -p server/web
	@rm -rf server/web/app.wasm
	@GOARCH=wasm GOOS=js go build -o server/web/app.wasm -ldflags $(LD_FLAGS) cmd/main.go

binary-arm: wasm
	@mkdir -p temp
	@GOOS=linux GOARM=7 GOARCH=arm go build -o temp/$(APP_NAME).arm -ldflags $(LD_FLAGS) cmd/main.go

binary: wasm
	@mkdir -p temp
	@go build -o temp/$(APP_NAME) -ldflags $(LD_FLAGS) cmd/main.go

deploy: binary-arm
	scp temp/$(APP_NAME).arm bb1:/tmp/$(APP_NAME)
	ssh bb1 /tmp/$(APP_NAME) -action deploy





