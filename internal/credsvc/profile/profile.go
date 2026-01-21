package profile

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type Name string

func (n Name) IsPseudo() bool {
	return strings.Contains(string(n), "/")
}

func Pseudo(accountID, roleName string) Name {
	return Name(fmt.Sprintf("%s/%s", accountID, roleName))
}

type List struct {
	Active   []Name
	Inactive []Name
}

func (l List) Sort() {
	slices.SortFunc(l.Active, compare)
	slices.SortFunc(l.Inactive, compare)
}

func compare(a, b Name) int {
	if a.IsPseudo() != b.IsPseudo() {
		return cmp.Compare(b, a)
	}

	return cmp.Compare(a, b)
}
