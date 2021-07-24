package main

import (
	"errors"
	"regexp"
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

type AzureLoadable struct {
	httpJson HttpJson
}

func AzureLoadableCreate() (*AzureLoadable, error) {
	return &AzureLoadable{
		httpJson: HttpJsonCreate(),
	}, nil
}

func (a *AzureLoadable) Load(ipv4Only bool) (Blocks, error) {

	// WARNING: hack incoming... Azure doesn't have a non-authenticated way to
	// grab its list of IP addresses via an api call, but you can visit a
	// webpage to get the "current" url of said data. This is downloading that
	// page and using a regex to pull out the URL we're after. Ugly, but it
	// beats requiring an Azure account for auth... Going to go with this until
	// it proves to flakey or something...
	url := "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"
	bodyBytes, err := a.httpJson.FetchBody(url, "GET")
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes[:])

	r := regexp.MustCompile(`click here to download.*href="(?P<url>[^"]+)"`)
	matches := r.FindStringSubmatch(body)
	if len(matches) != 2 {
		return nil, errors.New("Failed to find the download url (hacky)")
	}
	url = matches[1]

	values := azureValues{}
	err = a.httpJson.Fetch(url, "GET", &values)
	if err != nil {
		return nil, err
	}

	blocks := make(Blocks, 0)
	for _, value := range values.Values {
		if value.Properties.SystemService == "" || value.Properties.Region == "" {
			// If we don't have those fields this is just garbage/duplicate data
			continue
		}
		info := "Azure/" + value.Properties.SystemService + "/" + value.Properties.Region
		for _, cidr := range value.Properties.AddressPrefixes {
			if strings.Index(cidr, ":") != -1 && ipv4Only {
				// Ipv6 addr and we aren't interested
				continue
			}
			block, err := BlockCreateWithCidr(&cidr, &info)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}
