package awskey

import (
	"time"

	"kevwargo/aws-prompt/internal/credsvc/profile"
)

type Info struct {
	AccountID  string
	Profile    profile.Name
	Expiration *time.Time
}
