FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
COPY resources ./resources
COPY NotEnoughUpdates-REPO ./NotEnoughUpdates-REPO
COPY .env* ./

RUN go mod tidy
RUN go build -o yard-backend .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/yard-backend .
COPY --from=builder /app/resources ./resources
COPY --from=builder /app/NotEnoughUpdates-REPO ./NotEnoughUpdates-REPO
COPY --from=builder /app/.env* ./

EXPOSE 8080

CMD ["./yard-backend"]
