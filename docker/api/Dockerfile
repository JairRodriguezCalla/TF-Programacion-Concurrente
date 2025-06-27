# ---------- Etapa de build ----------
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o apiserver ./api

# ---------- Etapa de producción ----------
FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/apiserver /app/apiserver
ENV REDIS_ADDR=redis:6379
EXPOSE 8080
CMD ["./apiserver"]  
