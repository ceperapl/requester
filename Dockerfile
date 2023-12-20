# build the server binary
FROM golang:1.18-alpine AS base-builder
WORKDIR /go/src/github.com/ceperapl/requester
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server -tags release

# copy the server binary from builder stage; run the server binary
FROM alpine:3.16
WORKDIR /bin
COPY --from=base-builder /go/src/github.com/ceperapl/requester/bin/server .
ENTRYPOINT ["server"]
