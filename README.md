# newsapi

## Description

A Simple API server backed by RabbitMQ, Elasticsearch, Redis, and PostgreSQL. The server contains two services, news service acts as API gateway, and storage service consumes messages from RabbitMQ and stores them to PostgresSQL and Elasticsearch. The API documentation can be found in `newsapi.yml`.

## How to Run

### Docker Compose

```bash
# Clone the project
git clone https://github.com/rogaps/newsapi.git

# Move to directory
cd newsapi

# Run docker compose to build and start containers
docker-compose up -d

# API call examples
curl -i -XPOST --header 'Content-Type: application/json' 'http://localhost:8080/news' -d '{
    "author":"John Doe",
    "body":"Lorem ipsum dolor sit amet, consectetur adipiscing elit."
}'

curl -i -XPOST --header 'Content-Type: application/json' "http://localhost:8080/news" -d '{
    "author":"John Doe",
    "body":"Lorem ipsum dolor sit amet."
}'

curl -i -XPOST --header 'Content-Type: application/json' "http://localhost:8080/news" -d '{
    "author":"John Doe",
    "body":"Lorem ipsum."
}'

curl -XGET --header 'Content-Type: application/json' 'http://localhost:8080/news?page=1&limit=10' | jq

# Tear down containers
docker-compose down --rmi local
```

## TODO

- Writing unit tests, mock tests, and integration tests
- Code refactoring and better interfacing
- Better loggings

## Acknoledgements

- Gorilla Mux for http routing
- Uber's dig for dependency injection
- redis-go for Redis related operations
- Gorm for ORM
- etc.