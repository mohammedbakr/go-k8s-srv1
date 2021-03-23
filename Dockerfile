FROM golang:alpine AS builder
WORKDIR /go/src/github.com/k8-proxy/icap-service1
COPY . .
RUN cd cmd \
    && env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o  icap-service1 .

FROM alpine
COPY --from=builder /go/src/github.com/k8-proxy/icap-service1/cmd/icap-service1 /bin/icap-service1

RUN apk update && apk add ca-certificates

ENTRYPOINT ["/bin/icap-service1"]
