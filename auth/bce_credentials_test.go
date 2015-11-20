package auth

import (
	"os"
	"testing"
)

const (
	accessKeyId     = "accessKeyId"
	secretAccessKey = "secretAccessKey"
)

func TestNewBceCredentials(t *testing.T) {
	credential := NewBceCredentials(accessKeyId, secretAccessKey)
	if credential.AccessKeyId != accessKeyId {
		t.Errorf("NewBceCredentials AccessKeyId NOT Right")
	}

	if credential.SecretAccessKey != secretAccessKey {
		t.Errorf("NewBceCredentials SecretAccessKey NOT Right")
	}
}

func TestNewBceCredentialsFromEnv(t *testing.T) {
	os.Setenv("ACCESS_KEY_ID", accessKeyId)
	os.Setenv("SECRET_ACCESS_KEY", secretAccessKey)
	credential := NewBceCredentials("", "")

	if credential.AccessKeyId != accessKeyId {
		t.Errorf("NewBceCredentials AccessKeyId NOT Right")
	}

	if credential.SecretAccessKey != secretAccessKey {
		t.Errorf("NewBceCredentials SecretAccessKey NOT Right")
	}

}
