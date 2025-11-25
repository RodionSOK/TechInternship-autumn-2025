FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/pr-reviewer-service ./cmd/server

COPY migrations ./migrations

EXPOSE 8080

CMD ["./bin/pr-reviewer-service"]