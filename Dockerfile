FROM golang:alpine AS builder
WORKDIR /root
ADD . .
ARG VERSION
WORKDIR /root/cmd/sync
RUN env CGO_ENABLED=0 go build -v -trimpath -ldflags "-s -w -X main.version=${VERSION}"
FROM scratch
COPY --from=builder /root/cmd/sync/sync /sync
ENTRYPOINT ["/sync"]