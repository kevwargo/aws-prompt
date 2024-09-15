package awskey

import (
	"time"
)

type Info struct {
	AccountID   string
	Profile     *string
	AssumedRole *string
	Expiration  *time.Time
}
