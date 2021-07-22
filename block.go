package main

import (
	"bytes"
	"net"
	"strings"
)

type Block struct {
	net   *net.IPNet
	value *string
}

func BlockCreate(net *net.IPNet, value *string) *Block {
	return &Block{
		net:   net,
		value: value,
	}
}

func BlockCreateWithCidr(cidr *string, value *string) (*Block, error) {
	_, net, err := net.ParseCIDR(*cidr)
	if err != nil {
		return nil, err
	}
	return &Block{
		net:   net,
		value: value,
	}, nil
}

type Blocks []*Block

func (a Blocks) Len() int {
	return len(a)
}

func (a Blocks) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Blocks) Less(i, j int) bool {
	c := bytes.Compare(a[i].net.IP, a[j].net.IP)
	if c == 0 {
		return strings.Compare(*a[i].value, *a[j].value) == -1
	}
	return c == -1
}
