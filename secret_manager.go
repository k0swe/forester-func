package kellog

import (
	"cloud.google.com/go/secretmanager/apiv1"
	"context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type SecretStore struct {
	ctx    context.Context
	client *secretmanager.Client
}

func NewSecretStore(ctx context.Context) SecretStore {
	client, _ := secretmanager.NewClient(ctx)
	return SecretStore{ctx, client}
}

func (s *SecretStore) fetchSecret(secretId string) (string, error) {
	versionName := "projects/" + projectID + "/secrets/" + secretId + "/versions/latest"
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: versionName,
	}
	resp, err := s.client.AccessSecretVersion(s.ctx, req)
	if err != nil {
		return "", err
	}
	return string(resp.GetPayload().GetData()), nil
}
