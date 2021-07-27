package main

import (
	"sort"
	"strings"
)

type oraclePublicIpRanges struct {
	Regions []struct {
		Region string `json:"region"`
		Cidrs  []struct {
			Cidr string   `json:"cidr"`
			Tags []string `json:"tags"`
		} `json:"cidrs"`
	} `json:"regions"`
}

type OracleOrigin struct {
	httpJson HttpJson
	Emitter
}

func OracleOriginCreate() (*OracleOrigin, error) {
	return &OracleOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "oracle",
		},
	}, nil
}

func (o *OracleOrigin) Run(ipv4Only bool) error {
	ranges := oraclePublicIpRanges{}
	err := o.httpJson.Fetch("https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json", "GET", &ranges)
	if err != nil {
		return err
	}

	blocks := make(Blocks, 0)

	for _, region := range ranges.Regions {
		for _, cidr := range region.Cidrs {
			// Multiple tags for a given cidr, best we can do is sort and comma seperate them...
			sort.Strings(cidr.Tags)
			value := "Oracle/" + region.Region + "/" + strings.Join(cidr.Tags, ",")
			block, err := BlockCreateWithCidr(&cidr.Cidr, &value)
			if err != nil {
				return err
			}
			blocks = append(blocks, block)
		}
	}

	sort.Sort(blocks)

	for _, block := range blocks {
		o.Emit(block)
	}

	return o.Done()
}
