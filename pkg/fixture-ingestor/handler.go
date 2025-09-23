package fixtureingestor

import "github.com/uptrace/bun"

const (
	defaultRetries = 10
)

func NewFixtureIngestorWorker(
	queueURL string,
	database *bun.DB,
	sqsClient SQSClient,
	stopChan chan struct{},
	maxRetries int,
) (*FixtureIngestorWorker, error) {
	if maxRetries == 0 {
		maxRetries = defaultRetries
	}

	return &FixtureIngestorWorker{
		QueueURL:   queueURL,
		Database:   database,
		SQSClient:  sqsClient,
		StopChan:   stopChan,
		maxRetries: maxRetries,
	}, nil

}

func (w *FixtureIngestorWorker) Start() {

}
