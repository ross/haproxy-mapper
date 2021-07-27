package main

import (
	"errors"
	"fmt"
	"net"
	"strings"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

type recordIsp struct {
	Asn int    `maxminddb:"autonomous_system_number"`
	Isp string `maxminddb:"isp"`
}

type MaxMindIspOrigin struct {
	Filename string
	db       *maxminddb.Reader
	asn      Emitter
	isp      Emitter
}

func MaxMindIspOriginCreate(filename string) (*MaxMindIspOrigin, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(db.Metadata.DatabaseType, "ISP") &&
		!strings.HasSuffix(db.Metadata.DatabaseType, "ASN") {
		return nil, errors.New("provided database is not MaxMind ISP or ASN")
	}

	return &MaxMindIspOrigin{
		Filename: filename,
		db:       db,
		asn: Emitter{
			id: "maxmind.asn",
		},
		isp: Emitter{
			id: "maxmind.isp",
		},
	}, nil
}

func (m *MaxMindIspOrigin) AddAsnReceiver(receiver Receiver) {
	m.asn.AddReceiver(receiver)
}

func (m *MaxMindIspOrigin) AddIspReceiver(receiver Receiver) {
	m.isp.AddReceiver(receiver)
}

func (m *MaxMindIspOrigin) Run(ipv4Only bool) error {
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

	for networks.Next() {
		record := recordIsp{}
		net, err := networks.Network(&record)
		if err != nil {
			return err
		}

		if record.Asn != 0 {
			asn := fmt.Sprintf("AS%d", record.Asn)

			err = m.asn.Emit(BlockCreate(net, &asn))
			if err != nil {
				return err
			}
		}

		if record.Isp != "" {
			err = m.isp.Emit(BlockCreate(net, &record.Isp))
			if err != nil {
				return err
			}
		}
	}

	m.asn.Done()
	m.isp.Done()

	return nil
}
