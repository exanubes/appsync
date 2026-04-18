package internal

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type AwsCredentialsProvider struct {
	config      aws.Config
	mutex       sync.Mutex
	initialized bool
}

func (provider *AwsCredentialsProvider) Load(ctx context.Context) (aws.Credentials, error) {
	provider.mutex.Lock()
	cfg, err := provider.get_config(ctx)
	provider.mutex.Unlock()

	if err != nil {
		return aws.Credentials{}, err
	}

	return cfg.Credentials.Retrieve(ctx)

}

func (provider *AwsCredentialsProvider) get_config(ctx context.Context) (aws.Config, error) {
	if provider.initialized {
		return provider.config, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)

	if err == nil {
		provider.config = cfg
		provider.initialized = true
	}

	return provider.config, err
}
