FROM golang:1.24.4-alpine as base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o api

EXPOSE 8001

CMD ["./build/api"]