FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/server

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/user-service .

EXPOSE 50051

CMD [ "./user-service" ]