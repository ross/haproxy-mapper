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

type Cities struct {
	networks *maxminddb.Networks
	record   recordCity
}

func NewCities(networks *maxminddb.Networks) (*Cities, error) {
	return &Cities{
		networks: networks,
		record:   recordCity{},
	}, nil
}

func (c *Cities) Next() (*Block, error) {

	if c.networks.Next() {
		net, err := c.networks.Network(&c.record)
		if err != nil {
			return nil, err
		}
		location := ""
		if len(c.record.Continent.Code) > 0 {
			if len(c.record.Country.ISOCode) > 0 {
				if city, ok := c.record.City.Names["en"]; ok && len(city) > 0 {
					location = c.record.Continent.Code + "-" + c.record.Country.ISOCode + "-" + city
				} else {
					location = c.record.Continent.Code + "-" + c.record.Country.ISOCode
				}
			} else {
				location = c.record.Continent.Code
			}
		}
		return BlockCreate(net, &location), nil
	}

	return nil, nil
}

type MaxMindCitySource struct {
	Filename string
	Ipv4Only bool
	db       *maxminddb.Reader
}

func MaxMindCitySourceCreate(filename string, ipv4Only bool) (*MaxMindCitySource, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}

	return &MaxMindCitySource{
		Filename: filename,
		Ipv4Only: ipv4Only,
		db:       db,
	}, nil
}

func (m *MaxMindCitySource) Cities() (*Cities, error) {
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
	return NewCities(networks)
}
