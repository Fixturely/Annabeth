package fixtureingestor

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/uptrace/bun"
)

type SQSClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type FixtureIngestorWorker struct {
	SQSQueueURL string
	Database    *bun.DB
	// LocalCache utils.LocalCacheInterface[int64]
	SQSClient  SQSClient
	StopChan   chan struct{}
	maxRetries int
	Processing bool
}
