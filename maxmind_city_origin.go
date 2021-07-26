package main

import (
	"net"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

type recordCity struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Continent struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"continent"`
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}

type MaxMindCityOrigin struct {
	Filename string
	db       *maxminddb.Reader
	Emitter
}

func MaxMindCityOriginCreate(filename string) (*MaxMindCityOrigin, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}

	return &MaxMindCityOrigin{
		Filename: filename,
		db:       db,
		Emitter: Emitter{
			id: "aws",
		},
	}, nil
}

func (m *MaxMindCityOrigin) Run(ipv4Only bool) error {
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

	record := recordCity{}
	for networks.Next() {
		net, err := networks.Network(&record)
		if err != nil {
			return err
		}
		location := ""
		if len(record.Continent.Code) > 0 {
			if len(record.Country.ISOCode) > 0 {
				if city, ok := record.City.Names["en"]; ok && len(city) > 0 {
					location = record.Continent.Code + "-" + record.Country.ISOCode + "-" + city
				} else {
					location = record.Continent.Code + "-" + record.Country.ISOCode
				}
			} else {
				location = record.Continent.Code
			}
		}
		m.Emit(BlockCreate(net, &location))
	}

	m.Done()

	return nil
}
