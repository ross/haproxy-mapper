package main

import (
	"errors"
	"regexp"
	"sort"
	"strings"
)

type azureProperties struct {
	Region          string   `json:"region"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
}

type azureValue struct {
	Properties azureProperties `json:"properties"`
}

type azureValues struct {
	Values []azureValue `json:"values"`
}

type AzureSource struct {
	Ipv4Only bool
	blocks   Blocks
	loaded   bool
	httpJson HttpJson
}

func AzureSourceCreate(ipv4Only bool) (*AzureSource, error) {
	return &AzureSource{
		Ipv4Only: ipv4Only,
		blocks:   make(Blocks, 0),
		loaded:   false,
		httpJson: HttpJsonCreate(),
	}, nil
}

func (a *AzureSource) load() error {
	a.loaded = true

	// WARNING: hack incoming... Azure doesn't have a non-authenticated way to
	// grab its list of IP addresses via an api call, but you can visit a
	// webpage to get the "current" url of said data. This is downloading that
	// page and using a regex to pull out the URL we're after. Ugly, but it
	// beats requiring an Azure account for auth... Going to go with this until
	// it proves to flakey or something...
	url := "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"
	bodyBytes, err := a.httpJson.FetchBody(url, "GET")
	if err != nil {
		return err
	}
	body := string(bodyBytes[:])

	r := regexp.MustCompile(`click here to download.*href="(?P<url>[^"]+)"`)
	matches := r.FindStringSubmatch(body)
	if len(matches) != 2 {
		return errors.New("Failed to find the download url (hacky)")
	}
	url = matches[1]

	values := azureValues{}
	err = a.httpJson.Fetch(url, "GET", &values)
	if err != nil {
		return err
	}

	for _, value := range values.Values {
		if value.Properties.SystemService == "" || value.Properties.Region == "" {
			// If we don't have those fields this is just garbage/duplicate data
			continue
		}
		info := "Azure/" + value.Properties.SystemService + "/" + value.Properties.Region
		for _, cidr := range value.Properties.AddressPrefixes {
			if strings.Index(cidr, ":") != -1 && a.Ipv4Only {
				// Ipv6 addr and we aren't interested
				continue
			}
			block, err := BlockCreateWithCidr(&cidr, &info)
			if err != nil {
				return err
			}
			a.blocks = append(a.blocks, block)
		}
	}

	sort.Sort(a.blocks)

	return nil
}

func (a *AzureSource) Next() (*Block, error) {
	if !a.loaded {
		err := a.load()
		if err != nil {
			return nil, err
		}
	}

	n := len(a.blocks)
	if n > 0 {
		block := a.blocks[0]
		a.blocks = a.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
