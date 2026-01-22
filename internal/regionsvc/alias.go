package regionsvc

import (
	"fmt"
	"regexp"
	"strings"
)

func Shorten(region string) string {
	for short, long := range directions {
		region = strings.ReplaceAll(region, long, short)
	}

	return strings.ReplaceAll(region, "-", "")
}

func Expand(alias string) string {
	m := regexAlias.FindStringSubmatch(alias)
	if m == nil {
		return alias
	}

	dir := m[2]
	if dir != "" {
		dir = directions[dir]
	} else {
		dir = directions[m[3]] + directions[m[4]]
	}

	return fmt.Sprintf("%s-%s-%s", m[1], dir, m[5])
}

var (
	directions = map[string]string{
		"n": "north",
		"s": "south",
		"e": "east",
		"w": "west",
		"c": "central",
	}

	regexAlias = regexp.MustCompile("^([a-z][a-z])(?:([nsewc])|([ns])([ew]))([0-9]+)$")
)
