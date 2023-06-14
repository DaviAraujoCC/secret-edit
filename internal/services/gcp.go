package services

import (
	"context"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

type gcpSMServiceInterface interface {
	ListSecrets() error
	GetSecretInfo(secretId string) error
	GetSecretData(secretId string) error
	CreateSecret(secretId string) error
	CreateSecretVersion(secretId string, data []byte) error
}

type gcpSMService struct {
	client    *secretmanager.Client
	projectID string
}

func ReturnGCPSMService(projectID string) (*gcpSMService, error) {
	client, err := createGCPSMService()
	if err != nil {
		return nil, err
	}
	return &gcpSMService{
		client:    client,
		projectID: projectID,
	}, nil
}

func createGCPSMService() (*secretmanager.Client, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}
