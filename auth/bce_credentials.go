package auth

import (
	"os"
)

type BceCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
}

func NewBceCredentials(accessKeyId string, secretAccessKey string) *BceCredentials {
	if accessKeyId == "" {
		accessKeyId = os.Getenv("ACCESS_KEY_ID")
		if accessKeyId == "" {
			panic("No accessKeyId!")
		}
	}

	if secretAccessKey == "" {
		secretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
		if secretAccessKey == "" {
			panic("No secretAccessKey!")
		}
	}

	return &BceCredentials{
		AccessKeyId:     accessKeyId,
		SecretAccessKey: secretAccessKey,
	}
}
