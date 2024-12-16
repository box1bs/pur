.SILENT:

build:
	go build -o ./.bin/pur.exe ./cmd/app/main.go

run: build
	./.bin/pur.exe

test:
	go test -v ./...