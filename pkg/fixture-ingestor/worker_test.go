package fixtureingestor

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSQSClient is a mock implementation of SQSClient
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func TestFixtureIngestorWorker(t *testing.T) {
	mockSQS := &MockSQSClient{}
	stopChan := make(chan struct{})
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	// Test
	worker, err := NewFixtureIngestorWorker(queueURL, nil, mockSQS, stopChan, 10)
	assert.NoError(t, err)

	// Assert
	assert.NotNil(t, worker)
	assert.Equal(t, queueURL, worker.SQSQueueURL)
	assert.Nil(t, worker.Database)
	assert.Same(t, mockSQS, worker.SQSClient)
	assert.Equal(t, stopChan, worker.StopChan)
	assert.Equal(t, 10, worker.maxRetries)
	assert.True(t, worker.Processing)
}

func TestWorkerStart(t *testing.T) {
	mockSQS := &MockSQSClient{}
	stopChan := make(chan struct{})
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	// Prepare a single SQS message
	msgID := "msg-1"
	receipt := "rh-1"
	body := "{\"Id\":1,\"SportId\":2,\"TeamId1\":3,\"TeamId2\":4,\"DateTime\":\"2025-01-01T00:00:00Z\",\"Details\":{}}"

	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{{
				MessageId:     &msgID,
				ReceiptHandle: &receipt,
				Body:          &body,
			}},
		}, nil).Once()

	mockSQS.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.DeleteMessageOutput{}, nil).Once()

	// Set up subsequent ReceiveMessage calls to return empty (no more messages)
	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil).Maybe()

	worker, err := NewFixtureIngestorWorker(queueURL, nil, mockSQS, stopChan, 5)
	assert.NoError(t, err)

	// Run the worker
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.Start()
	}()

	// Give it time to process the message
	time.Sleep(200 * time.Millisecond)

	// Signal shutdown by closing the stop channel
	close(stopChan)

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped successfully
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop in time")
	}

	// Verify the worker is no longer processing
	assert.False(t, worker.Processing)
	mockSQS.AssertExpectations(t)
}

func TestWorkerGracefulShutdown(t *testing.T) {
	mockSQS := &MockSQSClient{}
	stopChan := make(chan struct{})
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	// Allow ReceiveMessage to be called and return no messages during shutdown
	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil).Maybe()

	worker, err := NewFixtureIngestorWorker(queueURL, nil, mockSQS, stopChan, 5)
	assert.NoError(t, err)

	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.Start()
	}()

	// Give the worker a brief moment to start
	time.Sleep(10 * time.Millisecond)

	// Signal graceful shutdown immediately
	close(stopChan)

	// Wait for graceful shutdown
	select {
	case <-done:
		// Worker stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not shut down gracefully in time")
	}

	// Verify the worker stopped processing
	assert.False(t, worker.Processing)
}

// Additional test for handling multiple messages
func TestWorkerMultipleMessages(t *testing.T) {
	mockSQS := &MockSQSClient{}
	stopChan := make(chan struct{})
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	// Prepare multiple messages
	msg1ID := "msg-1"
	receipt1 := "rh-1"
	body1 := "{\"Id\":1,\"SportId\":2,\"TeamId1\":3,\"TeamId2\":4,\"DateTime\":\"2025-01-01T00:00:00Z\",\"Details\":{}}"

	msg2ID := "msg-2"
	receipt2 := "rh-2"
	body2 := "{\"Id\":2,\"SportId\":2,\"TeamId1\":5,\"TeamId2\":6,\"DateTime\":\"2025-01-02T00:00:00Z\",\"Details\":{}}"

	// First call returns message 1
	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{{
				MessageId:     &msg1ID,
				ReceiptHandle: &receipt1,
				Body:          &body1,
			}},
		}, nil).Once()

	// Second call returns message 2
	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{{
				MessageId:     &msg2ID,
				ReceiptHandle: &receipt2,
				Body:          &body2,
			}},
		}, nil).Once()

	// Subsequent calls return empty
	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil).Maybe()

	// Expect both messages to be deleted
	mockSQS.On("DeleteMessage", mock.Anything, mock.MatchedBy(func(params *sqs.DeleteMessageInput) bool {
		return *params.ReceiptHandle == receipt1
	}), mock.Anything).Return(&sqs.DeleteMessageOutput{}, nil).Once()

	mockSQS.On("DeleteMessage", mock.Anything, mock.MatchedBy(func(params *sqs.DeleteMessageInput) bool {
		return *params.ReceiptHandle == receipt2
	}), mock.Anything).Return(&sqs.DeleteMessageOutput{}, nil).Once()

	worker, err := NewFixtureIngestorWorker(queueURL, nil, mockSQS, stopChan, 5)
	assert.NoError(t, err)

	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.Start()
	}()

	// Give it time to process both messages
	time.Sleep(400 * time.Millisecond)

	// Signal shutdown
	close(stopChan)

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped successfully
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop in time")
	}

	mockSQS.AssertExpectations(t)
}
