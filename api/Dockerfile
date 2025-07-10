FROM golang:1.24.5-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY /src ./src

RUN go build -o api ./src

FROM debian:bookworm-slim

WORKDIR /root/

COPY --from=builder /app ./app

COPY sqlite/DigitalIdentity.db ./app

WORKDIR /root/app

EXPOSE 8080

CMD ["./api"]