package main

import (
	"sort"
	"strings"
)

type cloudflarePublicIpList struct {
	Addresses     []string `json:"addresses"`
	Ipv6Addresses []string `json:"ipv6_addresses"`
}

type CloudflareSource struct {
	Ipv4Only bool
	blocks   Blocks
	loaded   bool
	httpJson HttpJson
}

func CloudflareSourceCreate(ipv4Only bool) (*CloudflareSource, error) {
	return &CloudflareSource{
		Ipv4Only: ipv4Only,
		blocks:   make(Blocks, 0),
		loaded:   false,
		httpJson: HttpJsonCreate(),
	}, nil
}

func (c *CloudflareSource) loadUrl(url string) error {
	body, err := c.httpJson.FetchBody(url, "GET")
	if err != nil {
		return err
	}
	value := "Cloudflare"
	for _, cidr := range strings.Split(string(body), "\n") {
		if len(cidr) == 0 {
			continue
		}
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return err
		}
		c.blocks = append(c.blocks, block)
	}

	return nil
}

func (c *CloudflareSource) load() error {
	c.loaded = true

	err := c.loadUrl("https://www.cloudflare.com/ips-v4")
	if err != nil {
		return err
	}

	if !c.Ipv4Only {
		err := c.loadUrl("https://www.cloudflare.com/ips-v6")
		if err != nil {
			return err
		}
	}

	sort.Sort(c.blocks)

	return nil
}

func (c *CloudflareSource) Next() (*Block, error) {
	if !c.loaded {
		err := c.load()
		if err != nil {
			return nil, err
		}
	}

	n := len(c.blocks)
	if n > 0 {
		block := c.blocks[0]
		c.blocks = c.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
