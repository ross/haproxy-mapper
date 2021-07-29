package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

func fatalUsage(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	outdir := flag.String("outdir", "", "Output directory")
	includeIpv6 := flag.Bool("ipv6", true, "Process IPv6 data, -ipv6=false to disable")
	includeAll := flag.Bool("all", false, "Include All data")
	includeAws := flag.Bool("aws", false, "Include AWS data")
	includeAzure := flag.Bool("azure", false, "Include Azure data")
	includeCloudflare := flag.Bool("cloudflare", false, "Include Cloudflare data")
	includeFastly := flag.Bool("fastly", false, "Include Fastly data")
	includeGoogleCloud := flag.Bool("google-cloud", false, "Include Google Cloud data")
	includeOracle := flag.Bool("oracle", false, "Include Oracle data")
	includeProvider := flag.Bool("provider", false, "Include providers map")
	asnDb := flag.String("asn-db", "", "MaxMind ASN database file")
	includeAsn := flag.Bool("asn", false, "Include asn map, requires -asn-db or -isp-db")
	ispDb := flag.String("isp-db", "", "MaxMind ISP database file")
	includeIsp := flag.Bool("isp", false, "Include isp map, requires -isp-db")
	cityDb := flag.String("city-db", "", "MaxMind City database file")
	includeLocation := flag.Bool("location", false, "Include location map, requires -city-db")
	includeContinent := flag.Bool("continent", false, "Include continent map, requires -city-db")
	includeCountry := flag.Bool("country", false, "Include country map, requires -city-db")
	includeSubdivisions := flag.Bool("subdivisions", false, "Include subdivisions map, requires -city-db")
	includeCity := flag.Bool("city", false, "Include city map, requires -city-db")
	includeSpamhaus := flag.Bool("spamhaus", false, "Include spamhaus data")

	flag.Parse()

	if *outdir == "" {
		log.Fatal("missing required argument -outdir")
	}

	if *includeAll {
		*includeIpv6 = true
		*includeAll = true
		*includeAws = true
		*includeAzure = true
		*includeCloudflare = true
		*includeFastly = true
		*includeGoogleCloud = true
		*includeOracle = true
		*includeProvider = true
		*includeAsn = true
		// includeIsp is special and is only enabled if we have an ISP db, see below
		*includeLocation = true
		*includeContinent = true
		*includeCountry = true
		*includeSubdivisions = true
		*includeCity = true
		*includeSpamhaus = true
	}

	var wg sync.WaitGroup

	runnables := make([]Runnable, 0)

	var providerMerger *MergingProcessor
	if *includeProvider {
		providerMerger = MergingProcessorCreate("provider")

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

	// This section makes sure our isp/asn params "make sense"
	if *asnDb != "" && *ispDb != "" {
		fatalUsage("-asn-db and -isp-db both specified, use -isp-db")
	} else if *includeIsp && *ispDb == "" {
		fatalUsage("argument -isp enabled, but -isp-db not specified")
	} else if *includeAsn && *ispDb == "" && *asnDb == "" {
		fatalUsage("argument -asn enabled, but neither -asn-db nor -isp-db specified")
	} else if *ispDb != "" && !*includeIsp && !*includeAsn {
		fatalUsage("argument -isp-db provided, but no associated maps included")
	} else if *asnDb != "" && !*includeAsn {
		fatalUsage("argument -asn-db provided, but no associated maps included")
	}

	ispAsnDb := *ispDb
	if *asnDb != "" {
		ispAsnDb = *asnDb
	}

	if ispAsnDb != "" {
		isp, err := MaxMindIspOriginCreate(ispAsnDb)
		if err != nil {
			log.Fatal(err)
		}

		if *ispDb != "" && !isp.HaveIspData {
			fatalUsage("argument -isp-db specified, but database does not include ISP data")
		}

		if *includeAll && isp.HaveIspData {
			*includeIsp = true
		}

		if *includeAsn {
			ipToAsn, err := MapDestinationCreate(path.Join(*outdir, "ip_to_asn"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			isp.AddAsnReceiver(reducer)
			reducer.AddReceiver(ipToAsn)
		}

		if *includeIsp {
			ipToIsp, err := MapDestinationCreate(path.Join(*outdir, "ip_to_isp"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			isp.AddIspReceiver(reducer)
			reducer.AddReceiver(ipToIsp)
		}

		wg.Add(1)
		runnables = append(runnables, isp)
	}

	if *cityDb != "" {
		if !*includeCity && !*includeContinent && !*includeCountry && !*includeLocation && !*includeSubdivisions {
			fatalUsage("argument -city-db provided, but no associated maps included")
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
			reducer := CombiningProcessorCreate()
			city.AddCityReceiver(reducer)
			reducer.AddReceiver(ipToCity)
		}

		if *includeContinent {
			ipToContinent, err := MapDestinationCreate(path.Join(*outdir, "ip_to_continent"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			city.AddContinentReceiver(reducer)
			reducer.AddReceiver(ipToContinent)
		}

		if *includeCountry {
			ipToCountry, err := MapDestinationCreate(path.Join(*outdir, "ip_to_country"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			city.AddCountryReceiver(reducer)
			reducer.AddReceiver(ipToCountry)
		}

		if *includeLocation {
			ipToLocation, err := MapDestinationCreate(path.Join(*outdir, "ip_to_location"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			city.AddLocationReceiver(reducer)
			reducer.AddReceiver(ipToLocation)
		}

		if *includeSubdivisions {
			ipToSubdivisions, err := MapDestinationCreate(path.Join(*outdir, "ip_to_subdivisions"))
			if err != nil {
				log.Fatal(err)
			}
			reducer := CombiningProcessorCreate()
			city.AddSubdivisionsReceiver(reducer)
			reducer.AddReceiver(ipToSubdivisions)
		}

		wg.Add(1)
		runnables = append(runnables, city)
	} else if *includeCity {
		fatalUsage("argument -city enabled, but -city-db not specified")
	} else if *includeContinent {
		fatalUsage("argument -continent enabled, but -city-db not specified")
	} else if *includeCountry {
		fatalUsage("argument -country enabled, but -city-db not specified")
	} else if *includeLocation {
		fatalUsage("argument -location enabled, but -city-db not specified")
	} else if *includeSubdivisions {
		fatalUsage("argument -subdivisions enabled, but -city-db not specified")
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

	if len(runnables) == 0 {
		fatalUsage("No outputs specified, see " + os.Args[0] + " -h for help")
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
