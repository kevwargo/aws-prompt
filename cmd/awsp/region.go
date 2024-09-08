package awsp

import "strings"

var directions = []string{"north", "south", "east", "west", "central"}

func shortenRegion(region string) string {
	for _, d := range directions {
		region = strings.ReplaceAll(region, d, d[:1])
	}

	return strings.ReplaceAll(region, "-", "")
}
