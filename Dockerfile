FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/light-simulator ./cmd/server

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/light-simulator .
COPY web/ web/

RUN mkdir -p uploads

EXPOSE 8080

ENV ENVIRONMENT=production
ENV PORT=8080

ENTRYPOINT ["./light-simulator"]
