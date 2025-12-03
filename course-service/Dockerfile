FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o course-service ./cmd/server

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root

COPY --from=builder /app/course-service .

EXPOSE 50052

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:50052/health || exit 1

CMD [ "./course-service" ]