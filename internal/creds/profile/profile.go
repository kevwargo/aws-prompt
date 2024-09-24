package profile

import (
	"fmt"
	"strings"
)

type Name string

func (n Name) IsPseudo() bool {
	return strings.Contains(string(n), "/")
}

func Pseudo(accountID, roleName string) Name {
	return Name(fmt.Sprintf("%s/%s", accountID, roleName))
}
