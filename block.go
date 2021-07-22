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

func (b *Block) Less(other *Block) bool {
	if b.net.IP.To4() == nil {
		// We're v6
		if other.net.IP.To4() != nil {
			// They're v4, they're "less"
			return false
		}
		// both v6 stuff below will apply
	} else if other.net.IP.To4() == nil {
		// We're v4 and they're v6, we're less
		return true
	}

	c := bytes.Compare(b.net.IP, other.net.IP)
	if c == 0 {
		return strings.Compare(*b.value, *other.value) == -1
	}
	return c == -1
}

type Blocks []*Block

func (a Blocks) Len() int {
	return len(a)
}

func (a Blocks) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Blocks) Less(i, j int) bool {
	return a[i].Less(a[j])
}
