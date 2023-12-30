.SILENT:

staticlint_main = ./cmd/staticlint/main.go
server_main = ./cmd/shortener/main.go
client_main = ./cmd/client/main.go

build_version = v1.0.0
commit = $(shell git rev-parse HEAD)
ifeq ($(OS), Windows_NT)
	build_date = $(shell date /t)
else
	build_date = $(shell date -I)
endif
ldflags = "-X main.buildCommit=$(commit) -X main.buildDate=$(build_date) -X main.buildVersion=$(build_version)"

ifeq ($(OS), Windows_NT)
	staticlint_exe = ./bin/win/staticlint.exe
	server_exe = ./bin/win/server.exe
	client_exe = ./bin/win/client.exe
else
	staticlint_exe = ./bin/unix/staticlint
	server_exe = ./bin/unix/server
	client_exe = ./bin/unix/client
endif

build-lint:
	go build -o $(staticlint_exe) $(staticlint_main)

lint: build-lint
	$(staticlint_exe) ./...

build-server:
	go build -ldflags $(ldflags) -o $(server_exe) $(server_main)

server: build-server
	$(server_exe)

build-client:
	go build -o $(client_exe) $(client_main)

client: build-client
	$(client_exe)

test:
	go test ./... -race