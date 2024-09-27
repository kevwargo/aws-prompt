package creds

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/cache"
	"kevwargo/aws-prompt/internal/creds/profile"
)

func Get(name profile.Name) (creds aws.Credentials, err error) {
	cache.WithCache(func(c cache.Cache) error {
		cached, err := c.Get(name)
		if err != nil {
			return err
		}

		if cached != nil {
			creds = *cached
			return nil
		}

		creds, err = profile.Resolve(name)
		if err != nil {
			return err
		}

		return c.Store(cache.StoreRequest{
			Profile: name,
			Creds:   creds,
		})
	})

	return
}

func Describe(accessKeyID string) (info awskey.Info, err error) {
	cache.WithCache(func(c cache.Cache) error {
		info, err = c.Info(accessKeyID)
		return err
	})

	return
}

func Refresh(accessKeyID string) (creds aws.Credentials, err error) {
	cache.WithCache(func(c cache.Cache) error {
		info, err := c.Info(accessKeyID)
		if err != nil {
			return err
		}

		if info.Profile == "" || info.Profile.IsPseudo() {
			return errors.New("Current credentials cannot be refreshed")
		}

		creds, err = Get(info.Profile)
		return err
	})

	return
}
