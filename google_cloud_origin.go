package main

import (
	"sort"
)

type gcIpRanges struct {
	Prefixes []struct {
		Ipv4Prefix string `json:"ipv4Prefix"`
		Ipv6Prefix string `json:"ipv6Prefix"`
		Service    string `json:"service"`
		Scope      string `json:"scope"`
	} `json:"prefixes"`
}

type GoogleCloudOrigin struct {
	httpJson HttpJson
	Emitter
}

func GoogleCloudOriginCreate() (*GoogleCloudOrigin, error) {
	return &GoogleCloudOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "google_cloud",
		},
	}, nil
}

func (g *GoogleCloudOrigin) Run(ipv4Only bool) error {
	header := Header{
		general: `#
# IP to Google Cloud mapping
#
# https://www.gstatic.com/ipranges/cloud.json
#
# Note: All "service" values appear to be "Google Cloud"
#
`,
		columns: "# cidr GC/service/scope\n",
	}
	if err := g.Header(header); err != nil {
		return err
	}

	ranges := gcIpRanges{}
	err := g.httpJson.Fetch("https://www.gstatic.com/ipranges/cloud.json", "GET", &ranges)
	if err != nil {
		return err
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
			return err
		}
		blocks = append(blocks, block)
	}

	sort.Sort(blocks)

	for _, block := range blocks {
		g.Emit(block)
	}

	g.Done()

	return nil
}
