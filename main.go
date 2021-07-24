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
		log.Fatal(err)
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
		log.Fatal(err)
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

func ip_to_provider(outfile string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	aws, err := AwsLoadableCreate()
	if err != nil {
		log.Fatal(err)
	}

	azure, err := AzureLoadableCreate()
	if err != nil {
		log.Fatal(err)
	}

	cloudflare, err := CloudflareLoadableCreate()
	if err != nil {
		log.Fatal(err)
	}

	fastly, err := FastlyLoadableCreate()
	if err != nil {
		log.Fatal(err)
	}

	gc, err := GoogleCloudLoadableCreate(ipv4Only)
	if err != nil {
		log.Fatal(err)
	}

	mapp, err := MapCreate(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer mapp.Close()

	sorter := MergingProcessorCreate()
	sorter.AddSource(BlockSourceCreate(aws, ipv4Only))
	sorter.AddSource(BlockSourceCreate(azure, ipv4Only))
	sorter.AddSource(BlockSourceCreate(cloudflare, ipv4Only))
	sorter.AddSource(BlockSourceCreate(fastly, ipv4Only))
	sorter.AddSource(BlockSourceCreate(gc, ipv4Only))

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
	go ip_to_provider(path.Join(outdir, "ip_to_provider"), ipv4Only, &wg)
	wg.Add(1)

	wg.Wait()
}
