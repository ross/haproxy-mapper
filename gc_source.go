package main

import (
	"fmt"
	"sort"
)

type gcPrefix struct {
	Ipv4Prefix string `json:"ipv4Prefix"`
	Ipv6Prefix string `json:"ipv6Prefix"`
	Service    string `json:"service"`
	Scope      string `json:"scope"`
}

type gcIpRanges struct {
	Prefixes []gcPrefix `json:"prefixes"`
}

type GcSource struct {
	Ipv4Only bool
	blocks   Blocks
	loaded   bool
	httpJson HttpJson
}

func GcSourceCreate(ipv4Only bool) (*GcSource, error) {
	return &GcSource{
		Ipv4Only: ipv4Only,
		blocks:   make(Blocks, 0),
		loaded:   false,
		httpJson: HttpJsonCreate(),
	}, nil
}

func (g *GcSource) load() error {
	g.loaded = true

	ranges := gcIpRanges{}
	err := g.httpJson.fetch("https://www.gstatic.com/ipranges/cloud.json", "GET", &ranges)
	if err != nil {
		return err
	}

	// TODO: DRY up these for loops?
	for _, prefix := range ranges.Prefixes {
		value := fmt.Sprintf("GC/%s/%s", prefix.Service, prefix.Scope)
		cidr := prefix.Ipv4Prefix
		if cidr == "" {
			if g.Ipv4Only {
				// Only interested in v4
				continue
			}
			cidr = prefix.Ipv6Prefix
		}
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return err
		}
		g.blocks = append(g.blocks, block)
	}

	sort.Sort(g.blocks)

	return nil
}

func (g *GcSource) Next() (*Block, error) {
	if !g.loaded {
		err := g.load()
		if err != nil {
			return nil, err
		}
	}

	n := len(g.blocks)
	if n > 0 {
		block := g.blocks[0]
		g.blocks = g.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
