FROM golang:1.26-trixie AS build
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build -o /app/bin/api ./cmd/api
RUN go build -o /app/bin/migrate ./cmd/migrate

FROM debian:trixie-slim
WORKDIR /app

COPY --from=build /app/bin/api ./api
COPY --from=build /app/bin/migrate ./migrate
COPY --from=build /app/migrations ./migrations

EXPOSE 8081
CMD ["./api"]