package main

import (
	"sort"
)

type fastlyPublicIpList struct {
	Addresses     []string `json:"addresses"`
	Ipv6Addresses []string `json:"ipv6_addresses"`
}

type FastlySource struct {
	Ipv4Only bool
	blocks   Blocks
	loaded   bool
	httpJson HttpJson
}

func FastlySourceCreate(ipv4Only bool) (*FastlySource, error) {
	return &FastlySource{
		Ipv4Only: ipv4Only,
		blocks:   make(Blocks, 0),
		loaded:   false,
		httpJson: HttpJsonCreate(),
	}, nil
}

func (f *FastlySource) load() error {
	f.loaded = true

	ranges := fastlyPublicIpList{}
	err := f.httpJson.Fetch("https://api.fastly.com/public-ip-list", "GET", &ranges)
	if err != nil {
		return err
	}

	addresses := make([]string, 0)
	addresses = append(addresses, ranges.Addresses...)
	if !f.Ipv4Only {
		addresses = append(addresses, ranges.Ipv6Addresses...)
	}

	value := "Fastly"
	for _, cidr := range addresses {
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return err
		}
		f.blocks = append(f.blocks, block)
	}

	sort.Sort(f.blocks)

	return nil
}

func (f *FastlySource) Next() (*Block, error) {
	if !f.loaded {
		err := f.load()
		if err != nil {
			return nil, err
		}
	}

	n := len(f.blocks)
	if n > 0 {
		block := f.blocks[0]
		f.blocks = f.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
