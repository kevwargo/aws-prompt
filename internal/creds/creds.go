package creds

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/cache"
	"kevwargo/aws-prompt/internal/creds/profile"
)

func Get(name profile.Name) (aws.Credentials, error) {
	c, err := cache.Open()
	if err != nil {
		return aws.Credentials{}, err
	}
	defer c.Close()

	creds, err := c.Get(name)
	if err != nil {
		return aws.Credentials{}, err
	}

	if creds != nil {
		return *creds, nil
	}

	return resolveProfile(name, c)
}

func Describe(accessKeyID string) (awskey.Info, error) {
	c, err := cache.Open()
	if err != nil {
		return awskey.Info{}, err
	}
	defer c.Close()

	return c.Info(accessKeyID)
}

func Refresh(accessKeyID string) (aws.Credentials, error) {
	c, err := cache.Open()
	if err != nil {
		return aws.Credentials{}, err
	}
	defer c.Close()

	info, err := c.Info(accessKeyID)
	if err != nil {
		return aws.Credentials{}, err
	}

	if info.Profile == "" || info.Profile.IsPseudo() {
		return aws.Credentials{}, errors.New("Current credentials cannot be refreshed")
	}

	return resolveProfile(info.Profile, c)
}
