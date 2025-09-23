package utils

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	appCfg "fixture-ingestor/config"
)

var (
	sqsClient *sqs.Client
)

func init() {
	log.Println("initializing sqs")
	cfg := appCfg.GetConfig()
	sqsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithHTTPClient(awsHttp.NewBuildableClient().WithTimeout(10*time.Second)),
	)
	if err != nil {
		log.Println("failed to load default config: %v", err)
		return
	}
	// add the endpoint to the config if it's set (for localstack)
	if cfg.AwsEndpoint != "" {
		sqsCfg.BaseEndpoint = aws.String(cfg.AwsEndpoint)
	}
	sqsClient = sqs.NewFromConfig(sqsCfg)
}

func GetSQSClient() *sqs.Client {
	return sqsClient
}
