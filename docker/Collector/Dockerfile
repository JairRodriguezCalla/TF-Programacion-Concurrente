FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o collector ./scripts/collect

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/collector /app/collector
ENV REDIS_ADDR=redis:6379
VOLUME ["/results"]                    # CSV saldrá aquí
CMD ["./collector"]
