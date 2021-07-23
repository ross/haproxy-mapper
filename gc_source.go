package main

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
	httpJson HttpJson
}

func GcSourceCreate(ipv4Only bool) (*GcSource, error) {
	return &GcSource{
		httpJson: HttpJsonCreate(),
	}, nil
}

func (g *GcSource) Load(ipv4Only bool) (Blocks, error) {
	ranges := gcIpRanges{}
	err := g.httpJson.Fetch("https://www.gstatic.com/ipranges/cloud.json", "GET", &ranges)
	if err != nil {
		return nil, err
	}

	blocks := make(Blocks, 0)
	for _, prefix := range ranges.Prefixes {
		value := "GC/" + prefix.Service + "/" + prefix.Scope
		cidr := prefix.Ipv4Prefix
		if cidr == "" {
			if ipv4Only {
				// Only interested in v4
				continue
			}
			cidr = prefix.Ipv6Prefix
		}
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
