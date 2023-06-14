package services

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
)

func (sm *gcpSMService) ListSecrets() ([]*secretmanagerpb.Secret, error) {
	ctx := context.Background()
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: "projects/" + sm.projectID,
	}
	var secretList []*secretmanagerpb.Secret
	it := sm.client.ListSecrets(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return secretList, err
		}
		secretList = append(secretList, resp)
	}
	return secretList, nil
}

func (sm *gcpSMService) GetSecretInfo(secretId string) (*secretmanagerpb.Secret, error) {
	ctx := context.Background()

	secretPath := fmt.Sprintf("projects/%s/secrets/%s", sm.projectID, secretId)

	req := &secretmanagerpb.GetSecretRequest{
		Name: secretPath,
	}
	secret, err := sm.client.GetSecret(ctx, req)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (sm *gcpSMService) GetSecretVersions(secretId string) ([]*secretmanagerpb.SecretVersion, error) {
	ctx := context.Background()

	secretPath := fmt.Sprintf("projects/%s/secrets/%s", sm.projectID, secretId)

	req := &secretmanagerpb.ListSecretVersionsRequest{
		Parent: secretPath,
	}
	var secretVersionList []*secretmanagerpb.SecretVersion
	it := sm.client.ListSecretVersions(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return secretVersionList, err
		}
		secretVersionList = append(secretVersionList, resp)
	}
	return secretVersionList, nil
}

func (sm *gcpSMService) GetSecretData(secretId, secretVersion string) ([]byte, error) {
	ctx := context.Background()

	secretPath := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", sm.projectID, secretId, secretVersion)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}
	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return result.Payload.Data, nil
}

func (sm *gcpSMService) CreateSecretVersion(secretId string, data io.Reader) error {
	ctx := context.Background()
	secretPath := fmt.Sprintf("projects/%s/secrets/%s", sm.projectID, secretId)

	// read bytes from data
	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(dataBytes)},
	}
	_, err = sm.client.AddSecretVersion(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
