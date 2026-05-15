FROM golang:1.26.2-alpine AS builder

WORKDIR /src
ENV GOWORK=off

COPY go.work go.work.sum ./
COPY platform ./platform
COPY reporting ./reporting
COPY shared ./shared
COPY document ./document

WORKDIR /src/document
RUN go build -mod=mod -o /out/document-http-server ./cmd/http-server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/document-http-server /app/document-http-server
COPY document/migrations /app/document/migrations

CMD ["/app/document-http-server"]
