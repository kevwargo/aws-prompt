package regionsvc

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Region struct {
	Name     string
	LongName string
}

var Defaults = []Region{
	{Name: "ap-northeast-1", LongName: "Asia Pacific (Tokyo)"},
	{Name: "ap-northeast-2", LongName: "Asia Pacific (Seoul)"},
	{Name: "ap-northeast-3", LongName: "Asia Pacific (Osaka)"},
	{Name: "ap-south-1", LongName: "Asia Pacific (Mumbai)"},
	{Name: "ap-southeast-1", LongName: "Asia Pacific (Singapore)"},
	{Name: "ap-southeast-2", LongName: "Asia Pacific (Sydney)"},
	{Name: "ca-central-1", LongName: "Canada (Central)"},
	{Name: "eu-central-1", LongName: "Europe (Frankfurt)"},
	{Name: "eu-north-1", LongName: "Europe (Stockholm)"},
	{Name: "eu-west-1", LongName: "Europe (Ireland)"},
	{Name: "eu-west-2", LongName: "Europe (London)"},
	{Name: "eu-west-3", LongName: "Europe (Paris)"},
	{Name: "sa-east-1", LongName: "South America (Sao Paulo)"},
	{Name: "us-east-1", LongName: "US East (N. Virginia)"},
	{Name: "us-east-2", LongName: "US East (Ohio)"},
	{Name: "us-west-1", LongName: "US West (N. California)"},
	{Name: "us-west-2", LongName: "US West (Oregon)"},
}

func Shorten(region string) string {
	for _, d := range directions {
		region = strings.ReplaceAll(region, d, d[:1])
	}

	return strings.ReplaceAll(region, "-", "")
}

type Resolver struct {
	nameCache  map[string]string
	cacheMutex sync.Mutex
}

func NewResolver() *Resolver {
	s := Resolver{
		nameCache: make(map[string]string, len(Defaults)),
	}

	for _, r := range Defaults {
		s.nameCache[r.Name] = r.LongName
	}

	return &s
}

func (r *Resolver) List(ctx context.Context, awscfg aws.Config) ([]Region, error) {
	log.Print("running ec2:DescribeRegions")
	regionsResp, err := ec2.NewFromConfig(awscfg).DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}
	log.Printf("ec2:DescribeRegions = %d regions", len(regionsResp.Regions))

	var (
		ssmClient *ssm.Client
		regions   []Region
	)

	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	for _, region := range regionsResp.Regions {
		name := *region.RegionName

		if longName, ok := r.nameCache[name]; ok {
			regions = append(regions, Region{Name: name, LongName: longName})
			log.Printf("found in cache %s -> %q", name, longName)
			continue
		}

		if ssmClient == nil {
			ssmClient = ssm.NewFromConfig(awscfg)
		}

		log.Printf("resolving long name of %s", name)
		paramResp, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name: aws.String(fmt.Sprintf("/aws/service/global-infrastructure/regions/%s/longName", name)),
		})
		if err != nil {
			return nil, err
		}

		longName := *paramResp.Parameter.Value
		log.Printf("resolved %s -> %q", name, longName)
		r.nameCache[name] = longName
		regions = append(regions, Region{Name: name, LongName: longName})
	}

	slices.SortFunc(regions, func(a, b Region) int { return cmp.Compare(a.Name, b.Name) })

	return regions, nil
}

var directions = []string{"north", "south", "east", "west", "central"}
