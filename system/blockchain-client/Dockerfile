FROM golang:1.24.5-alpine as builder

WORKDIR /app

COPY blockchain-client/go* ./

RUN go mod download

COPY blockchain-client/src ./src

RUN go build -o blockchain-client ./src

FROM alpine

WORKDIR /root/

COPY --from=builder /app ./app

WORKDIR /root/app

# solana keypairs
COPY blockchain-client/*json ./

EXPOSE 8001

CMD ["./blockchain-client"]
