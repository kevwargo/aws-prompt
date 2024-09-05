package awskey

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
)

// the algorithm is from https://hackingthe.cloud/aws/enumeration/get-account-id-from-keys
func DecodeAccountID(keyID string) (string, error) {
	b, err := base32.StdEncoding.DecodeString(keyID[4:])
	if err != nil {
		return "", err
	}

	b = append([]byte{0, 0}, b[:6]...)

	x := (binary.BigEndian.Uint64(b) & 0x7fffffffff80) >> 7

	return fmt.Sprintf("%012d", x), nil
}
