# Dockerfile (backend)
FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/server

FROM alpine
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD ["./app"]
