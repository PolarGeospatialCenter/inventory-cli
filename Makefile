test:
	go test -cover ./cmd/...

linux:
	GOOS=linux go build -o bin/inventory-cli.linux .
