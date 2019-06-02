FROM golang:alpine

WORKDIR /go/src/github.com/PolarGeospatialCenter/inventory-cli

RUN apk add --no-cache git make curl

COPY . ./
RUN make linux

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=0 /go/src/github.com/PolarGeospatialCenter/inventory-cli/bin/inventory-cli.linux /bin/inventory-cli
ENTRYPOINT ["/bin/inventory-cli"]
