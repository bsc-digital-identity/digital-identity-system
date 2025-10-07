FROM golang:1.24.5-alpine AS builder

WORKDIR /app


RUN apk add --no-cache gcc musl-dev

COPY api/go* ./

COPY ../pkg ./../pkg

RUN go mod download

COPY api/src ./src
RUN go build -o api ./src

FROM alpine:3.20

WORKDIR /root/app



COPY --from=builder /app/api .


COPY api/config.json ./config.json

EXPOSE 8080

CMD ["./api"]
