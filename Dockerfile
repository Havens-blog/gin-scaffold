# Build the manager binary
FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /opt
# Copy the Go Modules manifests
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags "-w -s" -a -o manager cmd/web/main.go
RUN ls -la /opt
FROM alpine AS prod
WORKDIR /opt
COPY public/ public/
COPY storage/ storage/
COPY config/ config/
COPY --from=builder /opt/manager /opt/manager
EXPOSE 20201
ENTRYPOINT ["/opt/manager"]