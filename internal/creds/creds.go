package creds

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/cache"
	"kevwargo/aws-prompt/internal/creds/profile"
)

func Get(name profile.Name) (creds aws.Credentials, err error) {
	cached, err := cache.Default.Get(name)
	if err != nil {
		return aws.Credentials{}, err
	}

	if cached != nil {
		return *cached, nil
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

func Refresh(accessKeyID string) (creds aws.Credentials, err error) {
	info, err := cache.Default.Info(accessKeyID)
	if err != nil {
		return aws.Credentials{}, err
	}

	if info.Profile == "" || info.Profile.IsPseudo() {
		return aws.Credentials{}, errors.New("Current credentials cannot be refreshed")
	}

	return Get(info.Profile)
}
