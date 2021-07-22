package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type awsIpv4Prefix struct {
	IpPrefix           string `json:"ip_prefix"`
	NetworkBorderGroup string `json:"network_border_group"`
	Region             string `json:"region"`
	Service            string `json:"service"`
}

type awsIpv6Prefix struct {
	IpPrefix           string `json:"ipv6_prefix"`
	NetworkBorderGroup string `json:"network_border_group"`
	Region             string `json:"region"`
	Service            string `json:"service"`
}

type awsIpRanges struct {
	Ipv4Prefixes []awsIpv4Prefix `json:"prefixes"`
	Ipv6Prefixes []awsIpv6Prefix `json:"ipv6_prefixes"`
}

type AwsSource struct {
	Ipv4Only bool
	client   http.Client
	blocks   Blocks
	loaded   bool
}

func AwsSourceCreate(ipv4Only bool) (*AwsSource, error) {
	return &AwsSource{
		Ipv4Only: ipv4Only,
		client: http.Client{
			Timeout: time.Duration(10 * time.Second),
		},
		blocks: make(Blocks, 0),
		loaded: false,
	}, nil
}

func (a *AwsSource) fetch(url string, method string, out interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("user-agent", "haproxy-mapper")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &out)
}

func (a *AwsSource) load() error {
	a.loaded = true

	ranges := awsIpRanges{}
	a.fetch("https://ip-ranges.amazonaws.com/ip-ranges.json", "GET", &ranges)

	specificBlocks := make(map[string]*Block)
	genericBlocks := make(map[string]*Block)

	// TODO: DRY up these for loops?
	for _, prefix := range ranges.Ipv4Prefixes {
		value := fmt.Sprintf("AWS/%s/%s", prefix.Service, prefix.Region)
		block, err := BlockCreateWithCidr(&prefix.IpPrefix, &value)
		if err != nil {
			return err
		}
		if prefix.Service == "AMAZON" {
			// Treat AMAZON service specially, there's lots of duplicates with
			// a more sepcific service and then generic amazon we only want the
			// generic when it's not available in a specific
			genericBlocks[block.net.String()] = block
		} else {
			specificBlocks[block.net.String()] = block
		}
	}

	if !a.Ipv4Only {
		for _, prefix := range ranges.Ipv6Prefixes {
			value := fmt.Sprintf("AWS/%s/%s", prefix.Service, prefix.Region)
			block, err := BlockCreateWithCidr(&prefix.IpPrefix, &value)
			if err != nil {
				return err
			}
			if prefix.Service == "AMAZON" {
				// Treat AMAZON service specially, there's lots of duplicates with
				// a more sepcific service and then generic amazon we only want the
				// generic when it's not available in a specific
				genericBlocks[block.net.String()] = block
			} else {
				specificBlocks[block.net.String()] = block
			}
		}
	}

	for _, block := range specificBlocks {
		// We want all the specifics
		a.blocks = append(a.blocks, block)
	}
	for net, block := range genericBlocks {
		if _, ok := specificBlocks[net]; !ok {
			// And the generics that don't exist as specifics
			a.blocks = append(a.blocks, block)
		}
	}

	sort.Sort(a.blocks)

	return nil
}

func (a *AwsSource) Next() (*Block, error) {
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
