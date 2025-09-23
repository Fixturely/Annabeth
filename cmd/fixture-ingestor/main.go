package main

import (
	"fixture-ingestor/config"
	fixtureingestor "fixture-ingestor/pkg/fixture-ingestor"
	"fixture-ingestor/utils"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const NUM_WORKERS = 1

func main() {
	cfg := config.GetConfig()

	// Set up worker synchronization
	var wg sync.WaitGroup
	stopChan := make(chan struct{}, NUM_WORKERS)

	// Create and start the worker
	workers := make([]*fixtureingestor.FixtureIngestorWorker, NUM_WORKERS)

	for i := range NUM_WORKERS {
		worker, err := fixtureingestor.NewFixtureIngestorWorker(
			cfg.SQS.FixturesProcessorQueueURL,
			utils.GetDatabase(),
			utils.GetSQSClient(),
			stopChan,
			20,
		)
		if err != nil {
			log.Fatalf("Failed to create fixture ingestor worker: %v", err)
		}
		workers[i] = worker
		wg.Add(1)
		go func(worker *fixtureingestor.FixtureIngestorWorker, workerId int) {
			defer wg.Done()
			worker.Start()
		}(worker, i)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down fixture ingestor workers...")

	// Stop all workers
	for range NUM_WORKERS {
		stopChan <- struct{}{}
	}

	// Wait for all workers to finish
	wg.Wait()
	log.Println("Fixture ingestor workers stopped")
}
