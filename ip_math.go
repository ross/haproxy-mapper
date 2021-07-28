package main

import (
	"net"
)

func NetLast(n *net.IPNet) *net.IP {
	ip := n.IP
	last := make(net.IP, len(ip))
	// Copy in the mask bits
	copy(last, n.Mask)
	for i := 0; i < len(last); i = i + 1 {
		// Invert the mask bit and | in the first/ip, mask will always be 0
		// anywhere IP is not
		last[i] = ^last[i] | ip[i]
	}

	return &last
}

func IpNext(ip *net.IP) *net.IP {
	// net.IP is annoying, net.IpV4 creates a 16-byte IP :-(, this function
	// doesn't handle it as the way IPs are generated by everything here does
	// 4-byte v4. If/when that changes this will need some updating
	n := len(*ip)
	out := make(net.IP, n)

	// Starting with a carry will be our +1
	carry := byte(1)
	// For each byte of the IP, working last to first
	for i := n - 1; i >= 0; i = i - 1 {
		// Each byte of the IP plus any carry from the previous
		out[i] = (*ip)[i] + carry
		if out[i] < (*ip)[i] {
			// Wrapped so we carry
			carry = 1
		} else {
			carry = 0
		}
	}

	return &out
}
