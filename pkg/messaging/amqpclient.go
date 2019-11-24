package messaging

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
)

// AMQPClient represents AMQP client
type AMQPClient struct {
	conn *amqp.Connection
}

// Connect creates an AMQP client
func Connect(connectionString string) (*AMQPClient, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("failed to connect to broker: connection string is empty")
	}
	var conn *amqp.Connection
	var err error
	r := retrier.New(retrier.ConstantBackoff(5, time.Second), nil)

	err = r.Run(func() error {
		conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker: %s", err)
	}

	return &AMQPClient{conn}, nil
}

// PublishOnQueue publishes a message to a queue
func (client *AMQPClient) PublishOnQueue(body []byte, queueName string) error {
	if client.conn == nil {
		return fmt.Errorf("failed to publish to broker: no connection")
	}

	ch, err := client.conn.Channel() // Get a channel from the connection
	defer ch.Close()

	queue, err := ch.QueueDeclare( // Declare a queue that will be created if not exists with some args
		queueName, // our queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	// Publishes a message onto the queue.
	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body, // Our JSON body as []byte
		})

	log.Infof("A message was sent to queue %s: %s\n", queueName, body)
	return err
}

// SubscribeToQueue subscribes client to a queue
func (client *AMQPClient) SubscribeToQueue(queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error {
	ch, err := client.conn.Channel()
	client.failOnError(err, "Failed to open a channel")

	queue, err := ch.QueueDeclare(
		queueName, // name of the queue
		false,     // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	client.failOnError(err, "Failed to declare a Queue")

	msgs, err := ch.Consume(
		queue.Name,   // queue
		consumerName, // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	client.failOnError(err, "Failed to register a consumer")

	go consumeLoop(msgs, handlerFunc)
	return nil
}

// Close closes connectioin
func (client *AMQPClient) Close() {
	if client.conn != nil {
		client.conn.Close()
	}
}

func (client *AMQPClient) failOnError(err error, msg string) {
	if err != nil {
		log.Errorln("%s: %s", msg, err)
	}
}

func consumeLoop(deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
	for d := range deliveries {
		// Invoke the handlerFunc func we passed as parameter.
		handlerFunc(d)
	}
}
