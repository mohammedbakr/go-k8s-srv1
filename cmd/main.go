package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/k8-proxy/k8-go-comm/pkg/minio"
	"github.com/k8-proxy/k8-go-comm/pkg/rabbitmq"
	"github.com/streadway/amqp"

	miniov7 "github.com/minio/minio-go/v7"
)

var (
	AdpatationReuquestExchange   = "adaptation-exchange"
	AdpatationReuquestRoutingKey = "adaptation-request"
	AdpatationReuquestQueueName  = "adaptation-request-queue"

	ProcessingRequestExchange   = "processing-request-exchange"
	ProcessingRequestRoutingKey = "processing-request"
	ProcessingRequestQueueName  = "processing-request-queue"

	inputMount                     = os.Getenv("INPUT_MOUNT")
	adaptationRequestQueueHostname = os.Getenv("ADAPTATION_REQUEST_QUEUE_HOSTNAME")
	adaptationRequestQueuePort     = os.Getenv("ADAPTATION_REQUEST_QUEUE_PORT")
	messagebrokeruser              = os.Getenv("MESSAGE_BROKER_USER")
	messagebrokerpassword          = os.Getenv("MESSAGE_BROKER_PASSWORD")

	minioEndpoint     = os.Getenv("MINIO_ENDPOINT")
	minioAccessKey    = os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey    = os.Getenv("MINIO_SECRET_KEY")
	sourceMinioBucket = os.Getenv("MINIO_SOURCE_BUCKET")
	cleanMinioBucket  = os.Getenv("MINIO_CLEAN_BUCKET")

	publisher   *amqp.Channel
	minioClient *miniov7.Client
)

func main() {

	// Get a connection
	connection, err := rabbitmq.NewInstance(adaptationRequestQueueHostname, adaptationRequestQueuePort, messagebrokeruser, messagebrokerpassword)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Initiate a publisher on processing exchange
	publisher, err = rabbitmq.NewQueuePublisher(connection, ProcessingRequestExchange)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer publisher.Close()

	// Start a consumer
	msgs, ch, err := rabbitmq.NewQueueConsumer(connection, AdpatationReuquestQueueName, AdpatationReuquestExchange, AdpatationReuquestRoutingKey)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer ch.Close()

	minioClient, err = minio.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey, false)

	if err != nil {
		log.Fatalf("%s", err)
	}

	exist, err := minio.CheckIfBucketExists(minioClient, sourceMinioBucket)
	if err != nil {
		log.Println("error minio connection", err)
		return
	}
	if !exist {

		err := minio.CreateNewBucket(minioClient, sourceMinioBucket)
		if err != nil {
			log.Println("error creating source  minio bucket ")
			return
		}
	}

	exist, err = minio.CheckIfBucketExists(minioClient, cleanMinioBucket)
	if err != nil {
		log.Println("error minio connection", err)
		return
	}
	if !exist {

		err := minio.CreateNewBucket(minioClient, cleanMinioBucket)
		if err != nil {
			log.Println("error creating clean minio bucket")
			return
		}
	}

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
	sourcePresignedURL, err := minio.UploadAndReturnURL(minioClient, sourceMinioBucket, input, time.Second*60*60*24)
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Printf("File uploaded to minio successfully: %s", sourcePresignedURL.String())
	d.Headers["source-presigned-url"] = sourcePresignedURL.String()
	d.Headers["reply-to"] = d.ReplyTo

	// Publish the details to Rabbit
	err = rabbitmq.PublishMessage(publisher, ProcessingRequestExchange, ProcessingRequestRoutingKey, d.Headers, []byte(""))
	if err != nil {
		return err
	}
	log.Printf("Message published to the processing queue : exchange : %s , routing key : %s , source-presigned-url : %s", ProcessingRequestExchange, ProcessingRequestRoutingKey, sourcePresignedURL.String())

	return nil
}
