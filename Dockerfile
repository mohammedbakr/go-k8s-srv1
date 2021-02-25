FROM golang:alpine AS builder
WORKDIR /go/src/github.com/k8-proxy/icap-service1
COPY . .
RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o  icap-service1 .

FROM scratch
COPY --from=builder /go/src/github.com/k8-proxy/icap-service1/icap-service1 /bin/icap-service1

ENTRYPOINT ["/bin/icap-service1"]
