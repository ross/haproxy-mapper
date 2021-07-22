package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type awsPrefix struct {
	IpPrefix           string `json:"ip_prefix"`
	Ipv6Prefix         string `json:"ipv6_prefix"`
	NetworkBorderGroup string `json:"network_border_group"`
	Region             string `json:"region"`
	Service            string `json:"service"`
}

type awsIpRanges struct {
	Prefixes     []awsPrefix `json:prefixes`
	Ipv6Prefixes []awsPrefix `json:ipv6_prefixes`
}

// TODO: should we formalize this and return it everywhere rather than sep net and value?
type block struct {
	net   *net.IPNet
	value *string
}

type AwsSource struct {
	Ipv4Only bool
	client   http.Client
	blocks   []block
	loaded   bool
}

func AwsSourceCreate(ipv4Only bool) (*AwsSource, error) {
	return &AwsSource{
		Ipv4Only: ipv4Only,
		client:   http.Client{
			Timeout: time.Duration(10 * time.Second),
		},
		blocks:   make([]block, 0),
		loaded:   false,
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
	prefixes := ranges.Prefixes
	if !a.Ipv4Only {
		prefixes = append(prefixes, ranges.Ipv6Prefixes...)
	}
	for _, prefix := range prefixes {
		_, net, err := net.ParseCIDR(prefix.IpPrefix)
		if err != nil {
			return err
		}
		value := fmt.Sprintf("AWS/%s/%s", prefix.Service, prefix.Region)
		a.blocks = append(a.blocks, block{
			net:   net,
			value: &value,
		})
	}

	return nil
}

func (a *AwsSource) Next() (*net.IPNet, *string, error) {
	if !a.loaded {
		err := a.load()
		if err != nil {
			return nil, nil, err
		}
	}

	n := len(a.blocks)
	if n > 0 {
		block := a.blocks[0]
		a.blocks = a.blocks[1:n]
		return block.net, block.value, nil
	}

	return nil, nil, nil
}
