package main

import (
	"flag"
	"log"
	"path"
	"sync"
)

type Runnable interface {
	Run(ipv4Only bool) error
}

func ip_to_droplist(outdir string, ipv4Only bool, wg *sync.WaitGroup) {
	defer wg.Done()

	asn, err := SpamhausOriginCreate()
	if err != nil {
		log.Fatal(err)
	}
	ipToAsn, err := MapDestinationCreate(path.Join(outdir, "ip_to_droplist"))
	if err != nil {
		log.Fatal(err)
	}
	asn.AddReceiver(ipToAsn)

	asn.Run(ipv4Only)
}

func main() {
	outdir := flag.String("outdir", "", "Output directory")
	includeIpv6 := flag.Bool("ipv6", false, "Process IPv6 data")
	includeAws := flag.Bool("aws", false, "Include AWS data")
	includeAzure := flag.Bool("azure", false, "Include Azure data")
	includeCloudflare := flag.Bool("cloudflare", false, "Include Cloudflare data")
	includeFastly := flag.Bool("fastly", false, "Include Fastly data")
	includeGoogleCloud := flag.Bool("google-cloud", false, "Include Google Cloud data")
	includeOracle := flag.Bool("oracle", false, "Include Oracle data")
	includeProvider := flag.Bool("provider", false, "Include providers map")
	asnDb := flag.String("asn-db", "", "MaxMind ASN database file")
	cityDb := flag.String("city-db", "", "MaxMind City database file")
	includeLocation := flag.Bool("location", false, "Include location map")
	includeContinent := flag.Bool("continent", false, "Include continent map")
	includeCountry := flag.Bool("country", false, "Include country map")
	includeSubdivisions := flag.Bool("subdivisions", false, "Include subdivisions map")
	includeCity := flag.Bool("city", false, "Include city map")
	includeSpamhaus := flag.Bool("spamhaus", false, "Include spamhaus data")

	flag.Parse()

	if *outdir == "" {
		log.Fatal("missing required argument -outdir")
	}

	var wg sync.WaitGroup

	runnables := make([]Runnable, 0)

	var providerMerger *MergingProcessor
	if *includeProvider {
		providerMerger = MergingProcessorCreate("providerMerger")

		ipToProvider, err := MapDestinationCreate(path.Join(*outdir, "ip_to_provider"))
		if err != nil {
			log.Fatal(err)
		}
		providerMerger.AddReceiver(ipToProvider)
	}

	if *includeAws {
		aws, err := AwsOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToAws, err := MapDestinationCreate(path.Join(*outdir, "ip_to_aws"))
		if err != nil {
			log.Fatal(err)
		}
		aws.AddReceiver(ipToAws)
		if providerMerger != nil {
			aws.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, aws)
	}

	if *includeAzure {
		azure, err := AzureOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToAzure, err := MapDestinationCreate(path.Join(*outdir, "ip_to_azure"))
		if err != nil {
			log.Fatal(err)
		}
		azure.AddReceiver(ipToAzure)
		if providerMerger != nil {
			azure.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, azure)
	}

	if *includeCloudflare {
		cloudflare, err := CloudflareOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToCloudflare, err := MapDestinationCreate(path.Join(*outdir, "ip_to_cloudflare"))
		if err != nil {
			log.Fatal(err)
		}
		cloudflare.AddReceiver(ipToCloudflare)
		if providerMerger != nil {
			cloudflare.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, cloudflare)
	}

	if *includeFastly {
		fastly, err := FastlyOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToFastly, err := MapDestinationCreate(path.Join(*outdir, "ip_to_fastly"))
		if err != nil {
			log.Fatal(err)
		}
		fastly.AddReceiver(ipToFastly)
		if providerMerger != nil {
			fastly.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, fastly)
	}

	if *includeGoogleCloud {
		googleCloud, err := GoogleCloudOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToGoogleCloud, err := MapDestinationCreate(path.Join(*outdir, "ip_to_google_cloud"))
		if err != nil {
			log.Fatal(err)
		}
		googleCloud.AddReceiver(ipToGoogleCloud)
		if providerMerger != nil {
			googleCloud.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, googleCloud)
	}

	if *includeOracle {
		oracle, err := OracleOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToOracle, err := MapDestinationCreate(path.Join(*outdir, "ip_to_oracle"))
		if err != nil {
			log.Fatal(err)
		}
		oracle.AddReceiver(ipToOracle)
		if providerMerger != nil {
			oracle.AddReceiver(providerMerger)
		}

		wg.Add(1)
		runnables = append(runnables, oracle)
	}

	if *asnDb != "" {
		asn, err := MaxMindAsnOriginCreate(*asnDb)
		if err != nil {
			log.Fatal(err)
		}
		ipToAsn, err := MapDestinationCreate(path.Join(*outdir, "ip_to_asn"))
		if err != nil {
			log.Fatal(err)
		}
		asn.AddReceiver(ipToAsn)

		wg.Add(1)
		runnables = append(runnables, asn)
	}

	if *cityDb != "" {
		if !*includeCity && !*includeContinent && !*includeCountry && !*includeLocation && !*includeSubdivisions {
			log.Fatal("argument -city-db provided, but no associated maps included")
		}

		city, err := MaxMindCityOriginCreate(*cityDb)
		if err != nil {
			log.Fatal(err)
		}

		if *includeCity {
			ipToCity, err := MapDestinationCreate(path.Join(*outdir, "ip_to_city"))
			if err != nil {
				log.Fatal(err)
			}
			city.AddCityReceiver(ipToCity)
		}

		if *includeContinent {
			ipToContinent, err := MapDestinationCreate(path.Join(*outdir, "ip_to_continent"))
			if err != nil {
				log.Fatal(err)
			}
			city.AddContinentReceiver(ipToContinent)
		}

		if *includeCountry {
			ipToCountry, err := MapDestinationCreate(path.Join(*outdir, "ip_to_country"))
			if err != nil {
				log.Fatal(err)
			}
			city.AddCountryReceiver(ipToCountry)
		}

		if *includeLocation {
			ipToLocation, err := MapDestinationCreate(path.Join(*outdir, "ip_to_location"))
			if err != nil {
				log.Fatal(err)
			}
			city.AddLocationReceiver(ipToLocation)
		}

		if *includeSubdivisions {
			ipToSubdivisions, err := MapDestinationCreate(path.Join(*outdir, "ip_to_subdivisions"))
			if err != nil {
				log.Fatal(err)
			}
			city.AddSubdivisionsReceiver(ipToSubdivisions)
		}

		wg.Add(1)
		runnables = append(runnables, city)
	} else if *includeCity {
		log.Fatal("argument -city enabled, but -city-db not specified")
	} else if *includeContinent {
		log.Fatal("argument -continent enabled, but -city-db not specified")
	} else if *includeCountry {
		log.Fatal("argument -country enabled, but -city-db not specified")
	} else if *includeLocation {
		log.Fatal("argument -location enabled, but -city-db not specified")
	} else if *includeSubdivisions {
		log.Fatal("argument -subdivisions enabled, but -city-db not specified")
	}

	if *includeSpamhaus {
		spamhaus, err := SpamhausOriginCreate()
		if err != nil {
			log.Fatal(err)
		}
		ipToSpamhaus, err := MapDestinationCreate(path.Join(*outdir, "ip_to_spamhaus"))
		if err != nil {
			log.Fatal(err)
		}
		spamhaus.AddReceiver(ipToSpamhaus)

		wg.Add(1)
		runnables = append(runnables, spamhaus)
	}

	for _, runnable := range runnables {
		go func(r Runnable) {
			defer wg.Done()
			err := r.Run(!*includeIpv6)
			if err != nil {
				log.Fatal(err)
			}
		}(runnable)
	}

	wg.Wait()
}
