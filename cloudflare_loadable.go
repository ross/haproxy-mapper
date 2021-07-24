package main

import (
	"strings"
)

type cloudflarePublicIpList struct {
	Addresses     []string `json:"addresses"`
	Ipv6Addresses []string `json:"ipv6_addresses"`
}

type CloudflareLoadable struct {
	httpJson HttpJson
}

func CloudflareLoadableCreate() (*CloudflareLoadable, error) {
	return &CloudflareLoadable{
		httpJson: HttpJsonCreate(),
	}, nil
}

func (c *CloudflareLoadable) loadUrl(url string) (Blocks, error) {
	body, err := c.httpJson.FetchBody(url, "GET")
	if err != nil {
		return nil, err
	}
	blocks := make(Blocks, 0)
	value := "Cloudflare"
	for _, cidr := range strings.Split(string(body), "\n") {
		if len(cidr) == 0 {
			continue
		}
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (c *CloudflareLoadable) Load(ipv4Only bool) (Blocks, error) {

	blocks := make(Blocks, 0)
	v4, err := c.loadUrl("https://www.cloudflare.com/ips-v4")
	if err != nil {
		return nil, err
	}
	blocks = append(blocks, v4...)

	if !ipv4Only {
		v6, err := c.loadUrl("https://www.cloudflare.com/ips-v6")
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, v6...)
	}

	return blocks, nil
}
