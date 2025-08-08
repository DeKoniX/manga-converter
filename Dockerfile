FROM golang:1.24.5-alpine

WORKDIR /app
COPY . .

RUN go build -o manga-converter ./cmd/main.go

CMD ["./manga-converter"]