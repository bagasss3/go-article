FROM golang:1.23.4-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main main.go

EXPOSE 8000

CMD ["./main", "s"]
