package fixtureingestor

import (
	"encoding/json"
	"fmt"
	"time"

	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Fixture struct {
	Id       int64
	SportId  int64
	TeamId1  int64
	TeamId2  int64
	DateTime time.Time
	Details  json.RawMessage
}

func (w *FixtureIngestorWorker) ValidateFixture(fixture *Fixture) error {
	if fixture.Details == nil {
		return fmt.Errorf("fixture details are required")
	}
	return nil
}

func (w *FixtureIngestorWorker) handleOrder(msg *sqsTypes.Message) error {
	fixture := Fixture{}
	err := json.Unmarshal([]byte(*msg.Body), &fixture)
	if err != nil {
		return fmt.Errorf("error unmarshalling message: %v", err)
	}

	err = w.ValidateFixture(&fixture)
	if err != nil {
		return fmt.Errorf("error validating fixture: %v", err)
	}

	return nil
}
