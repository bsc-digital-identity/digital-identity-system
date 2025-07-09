FROM golang:1.24.5-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o api

FROM alpine

WORKDIR /root/

COPY --from=builder /app ./app

EXPOSE 8080

CMD ["./app/api"]