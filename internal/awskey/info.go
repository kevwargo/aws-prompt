package awskey

import (
	"time"
)

type Info struct {
	AccountID   string
	Profile     *string
	SessionName *string
	Expiration  *time.Time
}
