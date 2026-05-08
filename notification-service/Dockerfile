FROM golang:1.25.6-alpine3.22

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/notification-service/main.go


CMD ["./main"]