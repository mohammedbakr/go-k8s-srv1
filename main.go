package main

import (
	"fmt"
	"log"
	"os"

	"github.com/k8-proxy/k8-go-comm/pkg/rabbitmq"
	"github.com/streadway/amqp"
)

var (
	exchange   = "adaptation-exchange"
	routingKey = "adaptation-request"
	queueName  = "adaptation-request-queue"

	processing_exchange   = "processing-exchange"
	processing_routingKey = "processing-request"
	processing_queueName  = "processing-queue"

	inputMount                     = os.Getenv("INPUT_MOUNT")
	adaptationRequestQueueHostname = os.Getenv("ADAPTATION_REQUEST_QUEUE_HOSTNAME")
	adaptationRequestQueuePort     = os.Getenv("ADAPTATION_REQUEST_QUEUE_PORT")
	messagebrokeruser              = os.Getenv("MESSAGE_BROKER_USER")
	messagebrokerpassword          = os.Getenv("MESSAGE_BROKER_PASSWORD")

	minioEndpoint     = os.Getenv("MINIO_ENDPOINT")
	minioAccessKey    = os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey    = os.Getenv("MINIO_SECRET_KEY")
	sourceMinioBucket = os.Getenv("MINIO_SOURCE_BUCKET")

	publisher *amqp.Channel
)

func main() {

	// Get a connection
	connection, err := rabbitmq.NewInstance(adaptationRequestQueueHostname, adaptationRequestQueuePort, messagebrokeruser, messagebrokerpassword)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Initiate a publisher on processing exchange
	publisher, err := rabbitmq.NewQueuePublisher(connection, processing_exchange)

	// Start a consumer
	msgs, err := rabbitmq.NewQueueConsumer(connection, queueName, exchange, routingKey)

	forever := make(chan bool)

	// Consume
	go func() {
		for d := range msgs {
			err := processMessage(d)
			if err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func processMessage(d amqp.Delivery) error {

	if d.Headers["file-id"] == nil ||
		d.Headers["source-file-location"] == nil ||
		d.Headers["rebuilt-file-location"] == nil {
		return fmt.Errorf("Headers value is nil")
	}

	fileID := d.Headers["file-id"].(string)
	input := d.Headers["source-file-location"].(string)

	log.Printf("Received a message for file: %s", fileID)

	// Upload the source file to Minio and Get presigned URL
	sourcePresignedURL, err := minio.UploadAndReturnURL(client, sourceMinioBucket, input)
	if err != nil {
		return err
	}

	// Publish the details to Rabbit
	err = rabbitmq.PublishMessage(channel, processing_exchange, processing_routingKey, d.Headers, []byte(""))
	if err != nil {
		return err
	}

	return nil
}
