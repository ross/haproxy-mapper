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
	/*

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
	*/
}

func ip_to_asn(src, outfile string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	/*
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
	*/
}

func ip_to_provider(outdir string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	aws, err := AwsOriginCreate()
	if err != nil {
		log.Fatal(err)
	}
	ipToAws, err := MapDestinationCreate(path.Join(outdir, "ip_to_aws"))
	if err != nil {
		log.Fatal(err)
	}
	aws.AddReceiver(ipToAws)

	azure, err := AzureOriginCreate()
	if err != nil {
		log.Fatal(err)
	}
	ipToAzure, err := MapDestinationCreate(path.Join(outdir, "ip_to_azure"))
	if err != nil {
		log.Fatal(err)
	}
	azure.AddReceiver(ipToAzure)

	cloudflare, err := CloudflareOriginCreate()
	if err != nil {
		log.Fatal(err)
	}
	ipToCloudflare, err := MapDestinationCreate(path.Join(outdir, "ip_to_cloudflare"))
	if err != nil {
		log.Fatal(err)
	}
	cloudflare.AddReceiver(ipToCloudflare)

	fastly, err := FastlyOriginCreate()
	if err != nil {
		log.Fatal(err)
	}
	ipToFastly, err := MapDestinationCreate(path.Join(outdir, "ip_to_fastly"))
	if err != nil {
		log.Fatal(err)
	}
	fastly.AddReceiver(ipToFastly)

	gc, err := GoogleCloudOriginCreate(ipv4Only)
	if err != nil {
		log.Fatal(err)
	}
	ipToGoogleCloud, err := MapDestinationCreate(path.Join(outdir, "ip_to_google_cloud"))
	if err != nil {
		log.Fatal(err)
	}
	gc.AddReceiver(ipToGoogleCloud)

	merger := MergingProcessorCreate("provider")
	aws.AddReceiver(merger)
	azure.AddReceiver(merger)
	cloudflare.AddReceiver(merger)
	fastly.AddReceiver(merger)
	gc.AddReceiver(merger)

	ipToProvider, err := MapDestinationCreate(path.Join(outdir, "ip_to_provider"))
	if err != nil {
		log.Fatal(err)
	}
	merger.AddReceiver(ipToProvider)

	aws.Run(ipv4Only)
	azure.Run(ipv4Only)
	cloudflare.Run(ipv4Only)
	fastly.Run(ipv4Only)
	gc.Run(ipv4Only)
}

func main() {
	var wg sync.WaitGroup

	outdir := "/tmp/mapper"
	ipv4Only := false

	/*
		go ip_to_location("tmp/GeoLite2-City.mmdb", path.Join(outdir, "ip_to_location"), ipv4Only, &wg)
		wg.Add(1)
		go ip_to_asn("tmp/GeoLite2-ASN.mmdb", path.Join(outdir, "ip_to_asn"), ipv4Only, &wg)
		wg.Add(1)
	*/
	go ip_to_provider(outdir, ipv4Only, &wg)
	wg.Add(1)

	wg.Wait()
}
