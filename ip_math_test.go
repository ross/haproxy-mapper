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

func TestIPNext(t *testing.T) {
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
		next := IPNext(&test.ip)
		if bytes.Compare(test.expected, *next) != 0 {
			t.Errorf("%s: IPNext(%s) -> %s, expected %s", test.desc, test.ip.String(), next.String(), test.expected.String())
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

type firstLastTest struct {
	first    net.IP
	last     net.IP
	expected []string
}

func TestIPNetFromFirstLast(t *testing.T) {
	for _, test := range []firstLastTest{
		// Singles
		{
			first:    net.IP{1, 2, 3, 4},
			last:     net.IP{1, 2, 3, 4},
			expected: []string{"1.2.3.4/32"},
		},
		{
			first:    net.IP{1, 2, 3, 0},
			last:     net.IP{1, 2, 3, 1},
			expected: []string{"1.2.3.0/31"},
		},
		{
			first:    net.IP{1, 2, 3, 128},
			last:     net.IP{1, 2, 3, 255},
			expected: []string{"1.2.3.128/25"},
		},
		{
			first:    net.IP{1, 2, 3, 0},
			last:     net.IP{1, 2, 3, 255},
			expected: []string{"1.2.3.0/24"},
		},
		{
			first:    net.IP{1, 2, 0, 0},
			last:     net.IP{1, 2, 1, 255},
			expected: []string{"1.2.0.0/23"},
		},
		{
			first:    net.IP{1, 2, 0, 0},
			last:     net.IP{1, 2, 127, 255},
			expected: []string{"1.2.0.0/17"},
		},
		{
			first:    net.IP{1, 2, 0, 0},
			last:     net.IP{1, 2, 255, 255},
			expected: []string{"1.2.0.0/16"},
		},
		{
			first:    net.IP{1, 0, 0, 0},
			last:     net.IP{1, 1, 255, 255},
			expected: []string{"1.0.0.0/15"},
		},
		{
			first:    net.IP{1, 0, 0, 0},
			last:     net.IP{1, 127, 255, 255},
			expected: []string{"1.0.0.0/9"},
		},
		{
			first:    net.IP{1, 0, 0, 0},
			last:     net.IP{1, 255, 255, 255},
			expected: []string{"1.0.0.0/8"},
		},
		{
			first:    net.IP{2, 0, 0, 0},
			last:     net.IP{3, 255, 255, 255},
			expected: []string{"2.0.0.0/7"},
		},
		{
			first:    net.IP{0, 0, 0, 0},
			last:     net.IP{127, 255, 255, 255},
			expected: []string{"0.0.0.0/1"},
		},
		{
			first:    net.IP{0, 0, 0, 0},
			last:     net.IP{255, 255, 255, 255},
			expected: []string{"0.0.0.0/0"},
		},
		// Multiples
		{
			// Every possible
			first: net.IP{0, 0, 0, 0},
			last:  net.IP{255, 255, 255, 254},
			expected: []string{
				"0.0.0.0/1",
				"128.0.0.0/2",
				"192.0.0.0/3",
				"224.0.0.0/4",
				"240.0.0.0/5",
				"248.0.0.0/6",
				"252.0.0.0/7",
				"254.0.0.0/8",
				"255.0.0.0/9",
				"255.128.0.0/10",
				"255.192.0.0/11",
				"255.224.0.0/12",
				"255.240.0.0/13",
				"255.248.0.0/14",
				"255.252.0.0/15",
				"255.254.0.0/16",
				"255.255.0.0/17",
				"255.255.128.0/18",
				"255.255.192.0/19",
				"255.255.224.0/20",
				"255.255.240.0/21",
				"255.255.248.0/22",
				"255.255.252.0/23",
				"255.255.254.0/24",
				"255.255.255.0/25",
				"255.255.255.128/26",
				"255.255.255.192/27",
				"255.255.255.224/28",
				"255.255.255.240/29",
				"255.255.255.248/30",
				"255.255.255.252/31",
				"255.255.255.254/32",
			},
		},
		{
			// Small then big
			first: net.IP{1, 0, 255, 255},
			last:  net.IP{1, 1, 255, 255},
			expected: []string{
				"1.0.255.255/32",
				"1.1.0.0/16",
			},
		},
		{
			// No reduction
			// block.net=81.2.69.142/31, block.value=EU
			// block.net=81.2.69.144/28, block.value=EU
			// block.net=81.2.69.160/27, block.value=EU
			// block.net=81.2.69.192/28, block.value=EU
			first: net.IP{81, 2, 69, 142},
			last:  net.IP{81, 2, 69, 207},
			expected: []string{
				"81.2.69.142/31",
				"81.2.69.144/28",
				"81.2.69.160/27",
				"81.2.69.192/28",
			},
		},
	} {
		got := make([]*net.IPNet, 0)
		IPNetFromFirstLast(&test.first, &test.last, &got)
		if got == nil {
			t.Errorf("IPNetFromFirstLast(%s, %s) -> nil", test.first.String(), test.last.String())
			continue
		} else if len(got) != len(test.expected) {
			t.Errorf("len(IPNetFromFirstLast(%s, %s)) -> %d, expected %d", test.first.String(), test.last.String(), len(got), len(test.expected))
			continue
		}
		for i, cidr := range test.expected {
			_, expected, _ := net.ParseCIDR(cidr)
			if bytes.Compare(expected.IP, got[i].IP) != 0 {
				t.Errorf("IPNetFromFirstLast(%s, %s)[%d].IP -> %s, expected %s", test.first.String(), test.last.String(), i,
					got[i].IP.String(), expected.IP.String())
				continue
			}
			if bytes.Compare(expected.Mask, got[i].Mask) != 0 {
				t.Errorf("IPNetFromFirstLast(%s, %s)[%d].Mask -> %s, expected %s", test.first.String(), test.last.String(), i,
					got[i].Mask.String(), expected.Mask.String())
				continue
			}
		}
	}
}
