.SILENT:

staticlint_main = ./cmd/staticlint/main.go
server_main = ./cmd/shortener/main.go
client_main = ./cmd/client/main.go

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
	go build -o $(server_exe) $(server_main)

server: build-server
	$(server_exe)

build-client:
	go build -o $(client_exe) $(client_main)

client: build-client
	$(client_exe)

test:
	go test ./... -race