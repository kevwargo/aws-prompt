package awskey

import (
	"time"

	"kevwargo/aws-prompt/internal/creds/profile"
)

type Info struct {
	AccountID  string
	Profile    profile.Name
	Expiration *time.Time
}
