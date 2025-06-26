FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o pushjob ./scripts/push_job

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/pushjob /app/pushjob
ENV REDIS_ADDR=redis:6379
CMD ["./pushjob"]
