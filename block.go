package main

import (
	"net"
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
