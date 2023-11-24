.SILENT:

staticlint_main = ./cmd/staticlint/main.go

ifeq ($(OS), Windows_NT)
	staticlint_exe = ./bin/win/staticlint.exe
else
	staticlint_exe = ./bin/unix/staticlint
endif

build-lint:
	go build -o $(staticlint_exe) $(staticlint_main)

lint: build-lint
	$(staticlint_exe) ./...

test:
	go test ./... -race