FROM golang:1.13

WORKDIR /app

COPY . .

RUN go mod download

COPY ./scripts/wait-for-it.sh wait-for-it.sh
RUN chmod +x wait-for-it.sh

RUN go build -o newsservice ./cmd/newsservice

EXPOSE 8080

CMD ["./newsservice"]
