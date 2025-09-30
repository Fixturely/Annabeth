package fixtureingestor

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/uptrace/bun"
)

const (
	defaultRetries = 10
)

func NewFixtureIngestorWorker(
	sqsQueueURL string,
	database *bun.DB,
	sqsClient SQSClient,
	stopChan chan struct{},
	maxRetries int,
) (*FixtureIngestorWorker, error) {
	if maxRetries == 0 {
		maxRetries = defaultRetries
	}

	return &FixtureIngestorWorker{
		SQSQueueURL: sqsQueueURL,
		Database:    database,
		SQSClient:   sqsClient,
		StopChan:    stopChan,
		maxRetries:  maxRetries,
		Processing:  true,
	}, nil

}

func (w *FixtureIngestorWorker) Start() {
	receivedMessagesFailCount := 0
	for {
		select {
		case <-w.StopChan:
			if w.Processing {
				w.Processing = false
			}
			return
		default:
			if !w.Processing {
				log.Println("Fixture ingestor worker is not processing")
				return
			}
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			msg, err := w.SQSClient.ReceiveMessage(
				ctxWithTimeout,
				&sqs.ReceiveMessageInput{
					QueueUrl:            &w.SQSQueueURL,
					MaxNumberOfMessages: 1,
					WaitTimeSeconds:     5,
				},
			)
			cancel()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				receivedMessagesFailCount++
				time.Sleep(time.Duration(receivedMessagesFailCount) * 2 * time.Second)
				if receivedMessagesFailCount > w.maxRetries {
					log.Println("Too many receive message failures, exiting")
					return
				}
				continue
			}

			// reset fail count on successful message receive
			receivedMessagesFailCount = 0

			if msg == nil || len(msg.Messages) == 0 {
				// No messages available, continue polling
				continue
			}

			// Process 1 message at a time
			err = w.handleOrder(&msg.Messages[0])
			if err != nil {
				log.Printf("Error handling message: %v", err)
				continue
			}

			// Delete the message from the queue
			_, err = w.SQSClient.DeleteMessage(
				context.Background(),
				&sqs.DeleteMessageInput{
					QueueUrl:      &w.SQSQueueURL,
					ReceiptHandle: msg.Messages[0].ReceiptHandle,
				},
			)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
				continue
			}
		}
	}
}
