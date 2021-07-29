package main

type fastlyPublicIpList struct {
	Addresses     []string `json:"addresses"`
	Ipv6Addresses []string `json:"ipv6_addresses"`
}

type FastlyOrigin struct {
	httpJson HttpJson
	Emitter
}

func FastlyOriginCreate() (*FastlyOrigin, error) {
	return &FastlyOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "fastly",
		},
	}, nil
}

func (f *FastlyOrigin) Run(ipv4Only bool) error {
	header := Header{
		general: `#
# IP to Fastly mapping
#
# https://api.fastly.com/public-ip-list
#
`,
		columns: "# cidr Fastly\n",
	}
	if err := f.Header(header); err != nil {
		return err
	}

	ranges := fastlyPublicIpList{}
	err := f.httpJson.Fetch("https://api.fastly.com/public-ip-list", "GET", &ranges)
	if err != nil {
		return err
	}

	addresses := make([]string, 0)
	addresses = append(addresses, ranges.Addresses...)
	if !ipv4Only {
		addresses = append(addresses, ranges.Ipv6Addresses...)
	}

	value := "Fastly"
	for _, cidr := range addresses {
		block, err := BlockCreateWithCidr(&cidr, &value)
		if err != nil {
			return err
		}
		err = f.Emit(block)
		if err != nil {
			return err
		}
	}

	return f.Done()
}
