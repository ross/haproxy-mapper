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
	Filename  string
	db        *maxminddb.Reader
	city      Emitter
	continent Emitter
	country   Emitter
	location  Emitter
}

func MaxMindCityOriginCreate(filename string) (*MaxMindCityOrigin, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}

	return &MaxMindCityOrigin{
		Filename: filename,
		db:       db,
		city: Emitter{
			id: "city.continent",
		},
		continent: Emitter{
			id: "city.continent",
		},
		country: Emitter{
			id: "city.country",
		},
		location: Emitter{
			id: "city.location",
		},
	}, nil
}

func (m *MaxMindCityOrigin) AddCityReceiver(receiver Receiver) {
	m.city.AddReceiver(receiver)
}

func (m *MaxMindCityOrigin) AddContinentReceiver(receiver Receiver) {
	m.continent.AddReceiver(receiver)
}

func (m *MaxMindCityOrigin) AddCountryReceiver(receiver Receiver) {
	m.country.AddReceiver(receiver)
}

func (m *MaxMindCityOrigin) AddLocationReceiver(receiver Receiver) {
	m.location.AddReceiver(receiver)
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

	for networks.Next() {
		record := recordCity{}
		net, err := networks.Network(&record)
		if err != nil {
			return err
		}
		location := ""
		if len(record.Continent.Code) > 0 {
			if len(record.Country.ISOCode) > 0 {
				if city, ok := record.City.Names["en"]; ok && len(city) > 0 {
					location = record.Continent.Code + "-" + record.Country.ISOCode + "-" + city
					m.city.Emit(BlockCreate(net, &city))
				} else {
					location = record.Continent.Code + "-" + record.Country.ISOCode
				}
				m.country.Emit(BlockCreate(net, &record.Country.ISOCode))
			} else {
				location = record.Continent.Code
			}
			m.continent.Emit(BlockCreate(net, &record.Continent.Code))
			m.location.Emit(BlockCreate(net, &location))
		}
	}

	m.city.Done()
	m.continent.Done()
	m.country.Done()
	m.location.Done()

	return nil
}
