# syntax=docker/dockerfile:1.7
FROM golang:1.26.2-alpine AS builder

WORKDIR /src
ENV GOWORK=off

COPY go.work go.work.sum ./
COPY platform ./platform
COPY reporting ./reporting
COPY shared ./shared
COPY document ./document

WORKDIR /src/document
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go build -p=2 -mod=mod -o /out/document-http-server ./cmd/http-server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/document-http-server /app/document-http-server
COPY document/migrations /app/document/migrations

CMD ["/app/document-http-server"]
