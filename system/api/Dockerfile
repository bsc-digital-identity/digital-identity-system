FROM golang:1.24.5-alpine as builder

WORKDIR /app

# Install build dependencies for CGO + SQLite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Enable CGO (required by go-sqlite3)
ENV CGO_ENABLED=1

COPY api/go* ./

COPY ../pkg ./../pkg

RUN go mod download

COPY api/src ./src
RUN go build -o api ./src

FROM alpine:3.20

WORKDIR /root/app

# Install SQLite runtime
RUN apk add --no-cache sqlite

COPY --from=builder /app/api .
#COPY sqlite/DigitalIdentity.db .

COPY api/*json ./

EXPOSE 8080

CMD ["./api"]
