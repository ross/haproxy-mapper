package main

import (
	"log"
	"path"
	"sync"
)

type Source interface {
	Next() (*Block, error)
}

func ip_to_location(src, outfile string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	mm, err := MaxMindCitySourceCreate(src, ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(outfile)
	if err != nil {
		log.Fatal("Failed to open map for writing: ", err)
	}
	defer mapp.Close()

	cities, err := mm.Cities()
	if err != nil {
		log.Fatal(err)
	}

	err = mapp.Consume(cities)
	if err != nil {
		log.Fatal(err)
	}
}

func ip_to_asn(src, outfile string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	mm, err := MaxMindAsnSourceCreate(src, ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(outfile)
	if err != nil {
		log.Fatal("Failed to open map for writing: ", err)
	}
	defer mapp.Close()

	asns, err := mm.Asns()
	if err != nil {
		log.Fatal(err)
	}

	err = mapp.Consume(asns)
	if err != nil {
		log.Fatal(err)
	}
}

func ip_to_cloud(outfile string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	aws, err := AwsSourceCreate(ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(outfile)
	if err != nil {
		log.Fatal("Failed to open map for writing: ", err)
	}
	defer mapp.Close()

	sorter := SortingProcessorCreate()
	sorter.AddSource(aws)

	err = mapp.Consume(sorter)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup

	outdir := "/tmp/mapper"
	ipv4Only := false

	go ip_to_location("tmp/GeoLite2-City.mmdb", path.Join(outdir, "ip_to_location"), ipv4Only, &wg)
	wg.Add(1)
	go ip_to_asn("tmp/GeoLite2-ASN.mmdb", path.Join(outdir, "ip_to_asn"), ipv4Only, &wg)
	wg.Add(1)
	go ip_to_cloud(path.Join(outdir, "ip_to_cloud"), ipv4Only, &wg)
	wg.Add(1)

	wg.Wait()
}
