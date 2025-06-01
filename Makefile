run:
	go run ./cmd/server/main.go

dev:
	air

cli:
	go run ./cmd/cli/main.go

build-cli:
	go build -o bin/okarun ./cmd/cli/main.go

install-cli:
	go install ./cmd/cli
