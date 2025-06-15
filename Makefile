run:
	go run ./cmd/server/main.go

dev:
	air

cli:
	go run ./cmd/cli/main.go

build-cli:
	go build -o bin/cli/okarun ./cmd/cli/main.go

build-server:
	go build -o bin/server/okarun ./cmd/server/main.go

release-snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check

release:
	goreleaser release --clean
