FROM golang:1.24.5-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY /src ./src

RUN go build -o blockchain-client ./src

FROM alpine

WORKDIR /root/

COPY --from=builder /app ./app

WORKDIR /root/app

COPY identity_app-keypair.json ./

EXPOSE 8001

CMD ["./blockchain-client"]
