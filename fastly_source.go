package main

type fastlyPublicIpList struct {
	Addresses     []string `json:"addresses"`
	Ipv6Addresses []string `json:"ipv6_addresses"`
}

type FastlySource struct {
	httpJson HttpJson
}

func FastlySourceCreate() (*FastlySource, error) {
	return &FastlySource{
		httpJson: HttpJsonCreate(),
	}, nil
}

func (f *FastlySource) Load(ipv4Only bool) (Blocks, error) {
	ranges := fastlyPublicIpList{}
	err := f.httpJson.Fetch("https://api.fastly.com/public-ip-list", "GET", &ranges)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0)
	addresses = append(addresses, ranges.Addresses...)
	if !ipv4Only {
		addresses = append(addresses, ranges.Ipv6Addresses...)
	}

	blocks := make(Blocks, 0)
	value := "Fastly"
	for _, cidr := range addresses {
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
