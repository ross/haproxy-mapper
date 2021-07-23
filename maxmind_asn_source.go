package main

import (
	"fmt"
	"net"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

type recordAsn struct {
	Asn int `maxminddb:"autonomous_system_number"`
}

type Asns struct {
	networks *maxminddb.Networks
	record   recordAsn
}

func NewAsns(networks *maxminddb.Networks) (*Asns, error) {
	return &Asns{
		networks: networks,
		record:   recordAsn{},
	}, nil
}

func (c *Asns) Next() (*Block, error) {

	if c.networks.Next() {
		net, err := c.networks.Network(&c.record)
		if err != nil {
			return nil, err
		}
		asn := fmt.Sprintf("AS%d", c.record.Asn)

		return BlockCreate(net, &asn), nil
	}

	return nil, nil
}

type MaxMindAsnSource struct {
	Filename string
	Ipv4Only bool
	db       *maxminddb.Reader
}

func MaxMindAsnSourceCreate(filename string, ipv4Only bool) (*MaxMindAsnSource, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}

	return &MaxMindAsnSource{
		Filename: filename,
		Ipv4Only: ipv4Only,
		db:       db,
	}, nil
}

func (m *MaxMindAsnSource) Asns() (*Asns, error) {
	cidr := "::/0"
	if m.Ipv4Only {
		cidr = "0.0.0.0/0"
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	networks := m.db.NetworksWithin(network)
	if !m.Ipv4Only {
		// We only want Ipv4 addresses once, if we call this when iterating
		// over 0.0.0.0/8 we get nothing for some reason
		maxminddb.SkipAliasedNetworks(networks)
	}
	return NewAsns(networks)
}
