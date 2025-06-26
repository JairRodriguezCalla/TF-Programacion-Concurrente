# ---------- build stage ----------
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# solo compilamos el worker
RUN go build -o worker ./nodes/worker

# ---------- runtime stage ----------
FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/worker /app/worker

# variables de configuración
ENV WORKERS=1          
ENV DELAY_MS=0         
ENV REDIS_ADDR=redis:6379

EXPOSE 8000            
CMD ["./worker"]
