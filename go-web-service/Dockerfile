FROM golang:1.23-alpine AS build

RUN apk add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o city-api

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/city-api .

CMD ["./city-api"]
