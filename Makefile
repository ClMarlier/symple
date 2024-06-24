dev:
	air --build.cmd "go build -o bin/symple cmd/main.go" --build.bin "./bin/symple"

check:
	staticcheck ./...

cover:
	go test -coverprofile=cover.out -coverpkg=symple
	go tool cover -html="cover.out"
