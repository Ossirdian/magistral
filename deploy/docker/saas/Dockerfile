# syntax=docker/dockerfile:1.7
FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
ENV GOWORK=off

COPY go.work go.work.sum ./
COPY iam/go.mod iam/go.sum ./iam/
COPY platform/go.mod platform/go.sum ./platform/
COPY reporting/go.mod reporting/go.sum ./reporting/
COPY resources/go.mod resources/go.sum ./resources/
COPY saas/go.mod saas/go.sum ./saas/
COPY student/go.mod student/go.sum ./student/
COPY shared/go.mod shared/go.sum ./shared/

COPY iam ./iam
COPY platform ./platform
COPY reporting ./reporting
COPY resources ./resources
COPY saas ./saas
COPY student ./student
COPY shared ./shared

RUN --mount=type=cache,target=/go/pkg/mod cd iam && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd platform && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd reporting && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd shared && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd resources && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd saas && go mod download
RUN --mount=type=cache,target=/go/pkg/mod cd student && go mod download

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build cd saas && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=2 -mod=mod -o /app/app-saas ./cmd/http-server

FROM alpine:3.22.0

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/app-saas ./
COPY --from=builder /app/saas/migrations ./saas/migrations

USER appuser

EXPOSE 8081

ENTRYPOINT ["./app-saas"]
