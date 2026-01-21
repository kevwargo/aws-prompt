package regions

import "strings"

func Shorten(region string) string {
	for _, d := range directions {
		region = strings.ReplaceAll(region, d, d[:1])
	}

	return strings.ReplaceAll(region, "-", "")
}

var directions = []string{"north", "south", "east", "west", "central"}
