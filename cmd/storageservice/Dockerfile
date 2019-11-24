FROM golang:1.13

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o storageservice ./cmd/storageservice

CMD ["./storageservice"]