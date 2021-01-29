package kellog

import (
	"cloud.google.com/go/secretmanager/apiv1"
	"context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const lotwUsername = "lotw_username"
const lotwPassword = "lotw_password"
const qrzLogbookApiKey = "qrz_logbook_api_key"

type SecretStore struct {
	ctx    context.Context
	client *secretmanager.Client
}

func NewSecretStore(ctx context.Context) SecretStore {
	client, _ := secretmanager.NewClient(ctx)
	return SecretStore{ctx, client}
}

func makeSecretId(scope string, key string) string {
	return scope + "_" + key
}

func (s *SecretStore) FetchSecret(logbookId string, key string) (string, error) {
	secretId := makeSecretId(logbookId, key)
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

// Add a version to the given secret, possibly creating the secret first. Returns the version name,
// e.g. "/projects/*/secrets/*/versions/*".
func (s *SecretStore) SetSecret(logbookId string, key string, secretValue string) (string, error) {
	secretId := makeSecretId(logbookId, key)
	projectName := "projects/" + projectID
	secretName := projectName + "/secrets/" + secretId
	_, err := s.client.GetSecret(s.ctx, &secretmanagerpb.GetSecretRequest{Name: secretName})
	if err != nil {
		// assume the secret didn't exist and create it
		secretName, err = s.createSecret(projectName, secretId)
		if err != nil {
			return "", err
		}
	}
	return s.addSecretVersion(secretName, secretValue)
}

// Create a new secret with no versions. Returns the secret name, e.g. "/projects/*/secrets/*".
func (s *SecretStore) createSecret(projectName string, secretId string) (string, error) {
	createResp, err := s.client.CreateSecret(s.ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   projectName,
		SecretId: secretId,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	return createResp.Name, nil
}

// Add a version to the given secret. Returns the version name, e.g. "/projects/*/secrets/*/versions/*".
func (s *SecretStore) addSecretVersion(secretName string, secretValue string) (string, error) {
	versionResp, err := s.client.AddSecretVersion(s.ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	})
	if err != nil {
		return "", err
	}
	return versionResp.Name, nil
}
