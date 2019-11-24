FROM golang:1.13

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o newsservice ./cmd/newsservice

EXPOSE 8080

CMD ["./newsservice"]
