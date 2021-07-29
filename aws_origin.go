package main

import (
	"sort"
)

type awsIpPrefix struct {
	// Annoyingly the aws json has two different ip prefixes, with the v6 added
	// to a single field in one of them. This ~hack lets us have a single type
	// that works for both. We won't access the properties directly and instead
	// will use the method to grab the one that applies. That will allow the
	// code that deals with these to be DRY.
	Ipv4Prefix string `json:"ip_prefix"`
	Ipv6Prefix string `json:"ipv6_prefix"`
	Region     string `json:"region"`
	Service    string `json:"service"`
}

func (a *awsIpPrefix) IpPrefix() string {
	if len(a.Ipv4Prefix) == 0 {
		return a.Ipv6Prefix
	}
	return a.Ipv4Prefix
}

type awsIpRanges struct {
	Ipv4Prefixes []awsIpPrefix `json:"prefixes"`
	Ipv6Prefixes []awsIpPrefix `json:"ipv6_prefixes"`
}

type AwsOrigin struct {
	httpJson HttpJson
	Emitter
}

func AwsOriginCreate() (*AwsOrigin, error) {
	return &AwsOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "aws",
		},
	}, nil
}

func (a *AwsOrigin) Run(ipv4Only bool) error {
	header := Header{
		general: `#
# IP to AWS mapping
#
# https://ip-ranges.amazonaws.com/ip-ranges.json
#
`,
		columns: "# cidr AWS/service/region\n",
	}
	if err := a.Header(header); err != nil {
		return err
	}

	ranges := awsIpRanges{}
	err := a.httpJson.Fetch("https://ip-ranges.amazonaws.com/ip-ranges.json", "GET", &ranges)
	if err != nil {
		return err
	}

	specificBlocks := make(map[string]*Block)
	genericBlocks := make(map[string]*Block)

	prefixes := make([]awsIpPrefix, 0)
	prefixes = append(prefixes, ranges.Ipv4Prefixes...)
	if !ipv4Only {
		prefixes = append(prefixes, ranges.Ipv6Prefixes...)
	}
	for _, prefix := range prefixes {
		value := "AWS/" + prefix.Service + "/" + prefix.Region
		cidr := prefix.IpPrefix()
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return err
		}
		if prefix.Service == "AMAZON" {
			// Treat AMAZON service specially, there's lots of duplicates with
			// a more sepcific service and then generic amazon we only want the
			// generic when it's not available in a specific
			genericBlocks[block.net.String()] = block
		} else {
			specificBlocks[block.net.String()] = block
		}
	}

	blocks := make(Blocks, 0)
	for _, block := range specificBlocks {
		// We want all the specifics
		blocks = append(blocks, block)
	}
	for net, block := range genericBlocks {
		if _, ok := specificBlocks[net]; !ok {
			// And the generics that don't exist as specifics
			blocks = append(blocks, block)
		}
	}

	sort.Sort(blocks)

	for _, block := range blocks {
		if err := a.Emit(block); err != nil {
			return err
		}
	}

	a.Done()

	return nil
}
