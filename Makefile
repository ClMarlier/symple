.PHONY: dev
dev:
	@air --build.cmd "go build -o bin/symple cmd/main.go" --build.bin "./bin/symple"

.PHONY: check
check:
	@staticcheck ./...

.PHONY: test
test:
	@go test

.PHONY: cover
cover:
	@go test -coverprofile=cover.out -coverpkg=./... && go tool cover -html=cover.out && rm cover.out
