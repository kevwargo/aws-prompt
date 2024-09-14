package awskey

import "time"

type Info struct {
	Profile    string
	Expiration *time.Time
}
