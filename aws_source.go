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

type AwsSource struct {
	Ipv4Only bool
	blocks   Blocks
	loaded   bool
	httpJson HttpJson
}

func AwsSourceCreate(ipv4Only bool) (*AwsSource, error) {
	return &AwsSource{
		Ipv4Only: ipv4Only,
		blocks:   make(Blocks, 0),
		loaded:   false,
		httpJson: HttpJsonCreate(),
	}, nil
}

func (a *AwsSource) load() error {
	a.loaded = true

	ranges := awsIpRanges{}
	err := a.httpJson.Fetch("https://ip-ranges.amazonaws.com/ip-ranges.json", "GET", &ranges)
	if err != nil {
		return err
	}

	specificBlocks := make(map[string]*Block)
	genericBlocks := make(map[string]*Block)

	prefixes := make([]awsIpPrefix, 0)
	prefixes = append(prefixes, ranges.Ipv4Prefixes...)
	if !a.Ipv4Only {
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

	for _, block := range specificBlocks {
		// We want all the specifics
		a.blocks = append(a.blocks, block)
	}
	for net, block := range genericBlocks {
		if _, ok := specificBlocks[net]; !ok {
			// And the generics that don't exist as specifics
			a.blocks = append(a.blocks, block)
		}
	}

	sort.Sort(a.blocks)

	return nil
}

func (a *AwsSource) Next() (*Block, error) {
	if !a.loaded {
		err := a.load()
		if err != nil {
			return nil, err
		}
	}

	n := len(a.blocks)
	if n > 0 {
		block := a.blocks[0]
		a.blocks = a.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
