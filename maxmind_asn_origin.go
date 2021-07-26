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

type MaxMindAsnOrigin struct {
	Filename string
	db       *maxminddb.Reader
	Emitter
}

func MaxMindAsnOriginCreate(filename string) (*MaxMindAsnOrigin, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}

	return &MaxMindAsnOrigin{
		Filename: filename,
		db:       db,
		Emitter: Emitter{
			id: "aws",
		},
	}, nil
}

func (m *MaxMindAsnOrigin) Run(ipv4Only bool) error {
	cidr := "::/0"
	if ipv4Only {
		cidr = "0.0.0.0/0"
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	networks := m.db.NetworksWithin(network)
	if !ipv4Only {
		// We only want Ipv4 addresses once, if we call this when iterating
		// over 0.0.0.0/8 we get nothing for some reason
		maxminddb.SkipAliasedNetworks(networks)
	}

	record := recordAsn{}
	for networks.Next() {
		net, err := networks.Network(&record)
		if err != nil {
			return err
		}
		asn := fmt.Sprintf("AS%d", record.Asn)

		err = m.Emit(BlockCreate(net, &asn))
		if err != nil {
			return err
		}
	}

	m.Done()

	return nil
}
