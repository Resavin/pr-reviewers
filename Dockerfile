FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./main.go

FROM alpine:latest
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations
CMD ["./app"]
