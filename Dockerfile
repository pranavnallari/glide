FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o glide ./cmd/glide

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/glide .
COPY config.yaml .

EXPOSE 8080

ENTRYPOINT ["./glide"]