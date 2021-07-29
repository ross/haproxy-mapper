package main

import (
	"errors"
	"sort"
	"strings"
)

type SpamhausOrigin struct {
	httpJson HttpJson
	Emitter
}

func SpamhausOriginCreate() (*SpamhausOrigin, error) {
	return &SpamhausOrigin{
		httpJson: HttpJsonCreate(),
		Emitter: Emitter{
			id: "cloudflare",
		},
	}, nil
}

func (s *SpamhausOrigin) runUrl(url string, blocks *Blocks) error {
	body, err := s.httpJson.FetchBody(url, "GET")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(body), "\n") {
		if len(line) == 0 || line[0] == ';' {
			continue
		}
		pieces := strings.Split(line, " ; ")
		if len(pieces) != 2 {
			return errors.New("Unexpected droplist line")
		}
		value := "Spamhaus" + "/" + pieces[1]
		block, err := BlockCreateWithCidr(&pieces[0], &value)
		if err != nil {
			return err
		}
		*blocks = append(*blocks, block)
	}

	return nil
}

func (s *SpamhausOrigin) Run(ipv4Only bool) error {
	header := Header{
		general: `#
# IP to Spamhaus mapping
#
# https://www.spamhaus.org/drop/drop.lasso
# https://www.spamhaus.org/drop/edrop.lasso
# https://www.spamhaus.org/drop/dropv6.txt
#
`,
		columns: "# cidr Spamhaus/id\n",
	}
	if err := s.Header(header); err != nil {
		return err
	}

	blocks := make(Blocks, 0)
	err := s.runUrl("https://www.spamhaus.org/drop/drop.lasso", &blocks)
	if err != nil {
		return err
	}

	err = s.runUrl("https://www.spamhaus.org/drop/edrop.lasso", &blocks)
	if err != nil {
		return err
	}

	if !ipv4Only {
		err = s.runUrl("https://www.spamhaus.org/drop/dropv6.txt", &blocks)
		if err != nil {
			return err
		}
	}

	sort.Sort(blocks)

	for _, block := range blocks {
		s.Emit(block)
	}

	return s.Done()
}
