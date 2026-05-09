FROM golang:1.25.6-alpine3.22

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./main.go

EXPOSE 8080
CMD ["./main"]