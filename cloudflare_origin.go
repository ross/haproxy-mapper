package main

import (
	"strings"
)

type CloudflareOrigin struct {
	httpJson HttpJson
	Emitter
}

func CloudflareOriginCreate() (*CloudflareOrigin, error) {
	return &CloudflareOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "cloudflare",
		},
	}, nil
}

func (c *CloudflareOrigin) runUrl(url string) error {
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
		c.Emit(block)
	}

	return nil
}

func (c *CloudflareOrigin) Run(ipv4Only bool) error {
	header := Header{
		general: `#
# IP to Cloudflare mapping
#
# https://www.cloudflare.com/ips-v4
#
`,
		columns: "# cidr Cloudflare\n",
	}
	if err := c.Header(header); err != nil {
		return err
	}

	err := c.runUrl("https://www.cloudflare.com/ips-v4")
	if err != nil {
		return err
	}

	if !ipv4Only {
		err := c.runUrl("https://www.cloudflare.com/ips-v6")
		if err != nil {
			return err
		}
	}

	return c.Done()
}
