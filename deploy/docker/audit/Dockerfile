# syntax=docker/dockerfile:1.7
FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
ENV GOWORK=off

COPY audit/go.mod ./audit/
COPY platform/go.mod platform/go.sum ./platform/

COPY audit ./audit
COPY platform ./platform

RUN --mount=type=cache,target=/go/pkg/mod cd audit && go mod download
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build cd audit && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=2 -mod=mod -o /app/app-audit ./cmd/http-server

FROM alpine:3.22.0

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/app-audit ./

USER appuser

EXPOSE 8088

ENTRYPOINT ["./app-audit"]
