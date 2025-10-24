FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o vps-monitor cmd/server/main.go

EXPOSE 8080

CMD ["./vps-monitor"]