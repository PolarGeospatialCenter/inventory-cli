vendor: Gopkg.lock
	dep ensure

test: vendor
	go test -cover ./cmd/...

linux: vendor
	GOOS=linux go build -o bin/inventory-cli.linux .
