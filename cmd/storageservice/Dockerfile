FROM golang:1.13

WORKDIR /app

COPY . .

COPY ./scripts/wait-for-it.sh wait-for-it.sh
RUN chmod +x wait-for-it.sh

RUN go mod download

RUN go build -o storageservice ./cmd/storageservice

CMD ["./storageservice"]