FROM golang:1.24-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o main /app/cmd/AuthService/main.go

# Run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
COPY cmd/env/set_env.sh .
RUN apk add --no-cache postgresql-client
EXPOSE 8080 8080
ENTRYPOINT ["./main"]