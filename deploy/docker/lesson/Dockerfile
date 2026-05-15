FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
ENV GOWORK=off

COPY go.work go.work.sum ./
COPY lesson/go.mod lesson/go.sum ./lesson/
COPY iam/go.mod iam/go.sum ./iam/
COPY platform/go.mod platform/go.sum ./platform/
COPY reporting/go.mod reporting/go.sum ./reporting/
COPY resources/go.mod resources/go.sum ./resources/
COPY saas/go.mod saas/go.sum ./saas/
COPY student/go.mod student/go.sum ./student/
COPY shared/go.mod shared/go.sum ./shared/

COPY lesson ./lesson
COPY iam ./iam
COPY platform ./platform
COPY reporting ./reporting
COPY resources ./resources
COPY saas ./saas
COPY student ./student
COPY shared ./shared

RUN cd lesson && go mod download
RUN cd iam && go mod download
RUN cd platform && go mod download
RUN cd reporting && go mod download
RUN cd shared && go mod download
RUN cd resources && go mod download
RUN cd saas && go mod download
RUN cd student && go mod download

RUN cd lesson && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o /app/app-lesson ./cmd/http-server

FROM alpine:3.22.0

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/app-lesson ./
COPY --from=builder /app/lesson/migrations ./lesson/migrations

USER appuser

EXPOSE 8087

ENTRYPOINT ["./app-lesson"]
