package fixtureingestor

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/uptrace/bun"
)

type SQSClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageBatch(ctx context.Context, params *sqs.DeleteMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
}

type FixtureIngestorWorker struct {
	QueueURL string
	Database *bun.DB
	// LocalCache utils.LocalCacheInterface[int64]
	SQSClient  SQSClient
	StopChan   chan struct{}
	maxRetries int
	batchSize  int
}
