package main

import (
	"bytes"
	"net"
	"testing"
)

type nextTest struct {
	desc     string
	ip       net.IP
	expected net.IP
}

func TestIpNext(t *testing.T) {
	for _, test := range []nextTest{
		{
			desc:     "simple",
			ip:       net.IP{0, 0, 0, 0},
			expected: net.IP{0, 0, 0, 1},
		},
		{
			desc:     "carry 1",
			ip:       net.IP{0, 0, 0, 255},
			expected: net.IP{0, 0, 1, 0},
		},
		{
			desc:     "carry 2",
			ip:       net.IP{0, 0, 255, 255},
			expected: net.IP{0, 1, 0, 0},
		},
		{
			desc:     "carry 3",
			ip:       net.IP{0, 255, 255, 255},
			expected: net.IP{1, 0, 0, 0},
		},
		{
			desc:     "wrap around",
			ip:       net.IP{255, 255, 255, 255},
			expected: net.IP{0, 0, 0, 0},
		},
		{
			desc:     "carry 1 plus",
			ip:       net.IP{0, 0, 1, 255},
			expected: net.IP{0, 0, 2, 0},
		},
		{
			desc:     "carry 2 plus",
			ip:       net.IP{0, 1, 255, 255},
			expected: net.IP{0, 2, 0, 0},
		},
		{
			desc:     "carry 3 plus",
			ip:       net.IP{1, 255, 255, 255},
			expected: net.IP{2, 0, 0, 0},
		},
		{
			desc:     "realistic 1",
			ip:       net.IP{192, 168, 1, 127},
			expected: net.IP{192, 168, 1, 128},
		},
		{
			desc:     "realistic 2",
			ip:       net.IP{192, 168, 43, 255},
			expected: net.IP{192, 168, 44, 0},
		},
		{
			desc: "ipv6 1",
			ip: net.IP{
				2, 0, 0, 0xf,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			expected: net.IP{
				2, 0, 0, 0xf,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
		{
			desc: "ipv6 2",
			ip: net.IP{
				2, 0, 0, 0xf,
				0, 0, 0, 0,
				0, 0, 0, 0,
				255, 255, 255, 255,
			},
			expected: net.IP{
				2, 0, 0, 0xf,
				0, 0, 0, 0,
				0, 0, 0, 1,
				0, 0, 0, 0,
			},
		},
		{
			desc: "ipv6 3",
			ip: net.IP{
				2, 0, 0, 0xf,
				255, 255, 255, 255,
				255, 255, 255, 255,
				255, 255, 255, 255,
			},
			expected: net.IP{
				2, 0, 0, 0x10,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
		},
		{
			desc: "ipv6 wrap",
			ip: net.IP{
				255, 255, 255, 255,
				255, 255, 255, 255,
				255, 255, 255, 255,
				255, 255, 255, 255,
			},
			expected: net.IP{
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
		},
	} {
		next := IpNext(&test.ip)
		if bytes.Compare(test.expected, *next) != 0 {
			t.Errorf("%s: IpNext(%s) -> %s, expected %s", test.desc, test.ip.String(), next.String(), test.expected.String())
		}
	}
}

type lastTest struct {
	desc     string
	cidr     string
	expected net.IP
}

func TestNetLast(t *testing.T) {
	for _, test := range []lastTest{
		{
			desc:     "all ipv4",
			cidr:     "0.0.0.0/0",
			expected: net.IP{255, 255, 255, 255},
		},
		{
			desc:     "/32 1",
			cidr:     "1.2.3.0/32",
			expected: net.IP{1, 2, 3, 0},
		},
		{
			desc:     "/32 2",
			cidr:     "1.2.3.128/32",
			expected: net.IP{1, 2, 3, 128},
		},
		{
			desc:     "/32 3",
			cidr:     "1.2.3.255/32",
			expected: net.IP{1, 2, 3, 255},
		},
		{
			desc:     "/31 1",
			cidr:     "1.2.3.254/31",
			expected: net.IP{1, 2, 3, 255},
		},
		{
			desc:     "/31 2",
			cidr:     "1.2.3.252/31",
			expected: net.IP{1, 2, 3, 253},
		},
		{
			desc:     "/25 1",
			cidr:     "1.2.3.0/25",
			expected: net.IP{1, 2, 3, 127},
		},
		{
			desc:     "/25 2",
			cidr:     "1.2.3.128/25",
			expected: net.IP{1, 2, 3, 255},
		},
		{
			desc:     "/24 1",
			cidr:     "1.2.3.0/24",
			expected: net.IP{1, 2, 3, 255},
		},
		{
			desc:     "/23 1",
			cidr:     "1.2.0.0/23",
			expected: net.IP{1, 2, 1, 255},
		},
		{
			desc:     "/17 1",
			cidr:     "1.2.0.0/17",
			expected: net.IP{1, 2, 127, 255},
		},
		{
			desc:     "/17 2",
			cidr:     "1.2.128.0/17",
			expected: net.IP{1, 2, 255, 255},
		},
		{
			desc:     "/16 1",
			cidr:     "1.2.0.0/16",
			expected: net.IP{1, 2, 255, 255},
		},
		{
			desc:     "/15 1",
			cidr:     "1.0.0.0/15",
			expected: net.IP{1, 1, 255, 255},
		},
		{
			desc:     "/15 2",
			cidr:     "1.254.0.0/15",
			expected: net.IP{1, 255, 255, 255},
		},
		{
			desc:     "/8 1",
			cidr:     "42.0.0.0/8",
			expected: net.IP{42, 255, 255, 255},
		},
		{
			desc: "ipv6 /0",
			cidr: "::/0",
			expected: net.IP{
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
			},
		},
		{
			desc: "ipv6 /96",
			cidr: "c002::/96",
			expected: net.IP{
				0xc0, 0x02, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0xff, 0xff, 0xff, 0xff,
			},
		},
		{
			desc: "ipv6 /95",
			cidr: "c002::/95",
			expected: net.IP{
				0xc0, 0x02, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 1,
				0xff, 0xff, 0xff, 0xff,
			},
		},
	} {
		_, n, _ := net.ParseCIDR(test.cidr)
		last := NetLast(n)
		if bytes.Compare(test.expected, *last) != 0 {
			t.Errorf("%s: NetLast(%s) -> %s, expected %s", test.desc, test.cidr, last.String(), test.expected.String())
		}
	}
}
