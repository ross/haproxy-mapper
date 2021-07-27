package main

import (
	"errors"
	"net"
	"strings"

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
	Filename     string
	db           *maxminddb.Reader
	city         Emitter
	continent    Emitter
	country      Emitter
	location     Emitter
	subdivisions Emitter
}

func MaxMindCityOriginCreate(filename string) (*MaxMindCityOrigin, error) {
	db, err := maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(db.Metadata.DatabaseType, "City") {
		return nil, errors.New("provided database is not MaxMind City")
	}

	return &MaxMindCityOrigin{
		Filename: filename,
		db:       db,
		city: Emitter{
			id: "maxmind.continent",
		},
		continent: Emitter{
			id: "maxmind.continent",
		},
		country: Emitter{
			id: "maxmind.country",
		},
		location: Emitter{
			id: "maxmind.location",
		},
		subdivisions: Emitter{
			id: "maxmind.subdivisions",
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

func (m *MaxMindCityOrigin) AddSubdivisionsReceiver(receiver Receiver) {
	m.subdivisions.AddReceiver(receiver)
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
		if len(record.Continent.Code) > 0 {
			location := record.Continent.Code
			m.continent.Emit(BlockCreate(net, &record.Continent.Code))

			if len(record.Country.ISOCode) > 0 {
				location += "/" + record.Country.ISOCode
				m.country.Emit(BlockCreate(net, &record.Country.ISOCode))

				subs := make([]string, 0)
				for _, subdivision := range record.Subdivisions {
					// There can be multiple subdivisions...
					if subdivisionName, ok := subdivision.Names["en"]; ok {
						location += "/" + subdivisionName
						subs = append(subs, subdivisionName)
					}
				}
				if len(subs) > 0 {
					// We'll use a comma seperated list here as that will allow
					// header parsing to see each subdivision as a seperate
					// header value when there are multiple.
					subs := strings.Join(subs, ", ")
					m.subdivisions.Emit(BlockCreate(net, &subs))
				}

				if city, ok := record.City.Names["en"]; ok && len(city) > 0 {
					location += "/" + city
					m.city.Emit(BlockCreate(net, &city))
				}
			}
			m.location.Emit(BlockCreate(net, &location))
		}
	}

	m.city.Done()
	m.continent.Done()
	m.country.Done()
	m.location.Done()
	m.subdivisions.Done()

	return nil
}
