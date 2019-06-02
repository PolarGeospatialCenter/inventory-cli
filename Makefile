test:
	go test -cover ./cmd/...

linux:
	GOOS=linux GO111MODULE=on go build -o bin/inventory-cli.linux .
