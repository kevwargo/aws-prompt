package credsvc

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/credsvc/cache"
	"kevwargo/aws-prompt/internal/credsvc/profile"
)

const (
	EnvAWSRegion          = "AWS_DEFAULT_REGION"
	EnvAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	EnvAWSSessionToken    = "AWS_SESSION_TOKEN"
)

var Envs = []string{
	EnvAWSAccessKeyID,
	EnvAWSSecretAccessKey,
	EnvAWSSessionToken,
}

func Get(name profile.Name, noCache bool) (creds aws.Credentials, err error) {
	if !noCache {
		cached, err := cache.Default.Get(name)
		if err != nil {
			return aws.Credentials{}, err
		}

		if cached != nil {
			return *cached, nil
		}
	}

	creds, err = profile.Resolve(name)
	if err != nil {
		return creds, err
	}

	err = cache.Default.Store(name, creds)
	if err != nil {
		return aws.Credentials{}, err
	}

	return creds, nil
}

func Describe(accessKeyID string) (info awskey.Info, err error) {
	return cache.Default.Info(accessKeyID)
}

func Refresh(accessKeyID string, noCache bool) (creds aws.Credentials, err error) {
	info, err := cache.Default.Info(accessKeyID)
	if err != nil {
		return aws.Credentials{}, err
	}

	if info.Profile == "" || info.Profile.IsPseudo() {
		return aws.Credentials{}, errors.New("Current credentials cannot be refreshed")
	}

	return Get(info.Profile, noCache)
}

func Map(creds aws.Credentials) map[string]string {
	return map[string]string{
		EnvAWSAccessKeyID:     creds.AccessKeyID,
		EnvAWSSecretAccessKey: creds.SecretAccessKey,
		EnvAWSSessionToken:    creds.SessionToken,
	}
}
