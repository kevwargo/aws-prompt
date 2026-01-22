package regionsvc_test

import (
	"testing"

	"kevwargo/aws-prompt/internal/regionsvc"
)

func TestAliases(t *testing.T) {
	table := map[string]string{
		"afs1":  "af-south-1",
		"ape1":  "ap-east-1",
		"ape2":  "ap-east-2",
		"apne1": "ap-northeast-1",
		"apne2": "ap-northeast-2",
		"apne3": "ap-northeast-3",
		"aps1":  "ap-south-1",
		"aps2":  "ap-south-2",
		"apse1": "ap-southeast-1",
		"apse2": "ap-southeast-2",
		"apse3": "ap-southeast-3",
		"apse4": "ap-southeast-4",
		"apse5": "ap-southeast-5",
		"apse6": "ap-southeast-6",
		"apse7": "ap-southeast-7",
		"cac1":  "ca-central-1",
		"caw1":  "ca-west-1",
		"euc1":  "eu-central-1",
		"euc2":  "eu-central-2",
		"eun1":  "eu-north-1",
		"eus1":  "eu-south-1",
		"eus2":  "eu-south-2",
		"euw1":  "eu-west-1",
		"euw2":  "eu-west-2",
		"euw3":  "eu-west-3",
		"ilc1":  "il-central-1",
		"mec1":  "me-central-1",
		"mes1":  "me-south-1",
		"mxc1":  "mx-central-1",
		"sae1":  "sa-east-1",
		"use1":  "us-east-1",
		"use2":  "us-east-2",
		"usw1":  "us-west-1",
		"usw2":  "us-west-2",
	}

	for a, f := range table {
		alias := regionsvc.Shorten(f)
		full := regionsvc.Expand(a)
		if a != alias {
			t.Errorf("Shorten(%s) == %s, expected %s", f, alias, a)
		}
		if f != full {
			t.Errorf("Expand(%s) == %s, expected %s", a, full, f)
		}
	}
}
