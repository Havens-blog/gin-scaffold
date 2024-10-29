# Build the manager binary
FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /opt
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY vendor/ vendor/

# Copy the go source
COPY config/ config/
COPY public/ public/
COPY storage/ storage/
COPY cmd/web/main.go cmd/web/main.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags "-w -s" -a -o manager cmd/web/main.go

FROM alpine AS prod
COPY --from=builder /opt/manager .
EXPOSE 8002
ENTRYPOINT ["/manager"]