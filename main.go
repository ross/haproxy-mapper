package main

import (
	"log"
	"path"
	"sync"
)

func ip_to_location(src, outdir string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	mm, err := MaxMindCitySourceCreate(src, ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(path.Join(outdir, "ip_to_location"))
	if err != nil {
		log.Fatal("Failed to open map for writing: ", err)
	}
	defer mapp.Close()

	cities, err := mm.Cities()
	if err != nil {
		log.Fatal(err)
	}

	net, location, err := cities.Next()
	for ; net != nil && err == nil; net, location, err = cities.Next() {
		if len(*location) == 0 {
			continue
		}
		if err := mapp.Write(net, location); err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func ip_to_asn(src, outdir string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	mm, err := MaxMindAsnSourceCreate(src, ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(path.Join(outdir, "ip_to_asn"))
	if err != nil {
		log.Fatal("Failed to open map for writing: ", err)
	}
	defer mapp.Close()

	asns, err := mm.Asns()
	if err != nil {
		log.Fatal(err)
	}

	net, asn, err := asns.Next()
	for ; asn != nil && err == nil; net, asn, err = asns.Next() {
		if len(*asn) == 0 {
			continue
		}
		if err := mapp.Write(net, asn); err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup

	go ip_to_location("tmp/GeoLite2-City.mmdb", "/tmp/mapper", true, &wg)
	wg.Add(1)
	go ip_to_asn("tmp/GeoLite2-ASN.mmdb", "/tmp/mapper", true, &wg)
	wg.Add(1)

	wg.Wait()
}
