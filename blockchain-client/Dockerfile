FROM golang:1.24.5-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o blockchain-client

FROM alpine

WORKDIR /root/

COPY --from=builder /app ./app

EXPOSE 8001

CMD ["./app/blockchain-client"]
