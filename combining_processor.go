package main

import (
	"bytes"
	"log"
	"net"
)

type CombiningProcessor struct {
	first    net.IP
	last     net.IP
	next     net.IP
	value    string
	heldBack *Block
	Emitter
}

func CombiningProcessorCreate() *CombiningProcessor {
	return &CombiningProcessor{
		first: net.IP{},
		last:  net.IP{},
		value: "",
	}
}

func (m *CombiningProcessor) Subscribed(id string) {
}

func (m *CombiningProcessor) Header(id string, header Header) error {
	header.general += "# Adjacent blocks with matching have been combined\n#\n"
	return m.Emitter.Header(header)
}

func (m *CombiningProcessor) buildAndEmit() error {
	log.Printf("buildAndEmit: first=%s, last=%s, value='%s'", m.first.String(), m.last.String(), m.value)

	num_bits := 32
	if m.first.To4() == nil {
		num_bits = 128
	}

	bit := 0
	for bit < num_bits {
		mask := net.CIDRMask(bit, num_bits)
		log.Printf("mask=%d (%s)", bit, mask)
		candidate := &net.IPNet{m.first, mask}
		candidate_last := NetLast(candidate)
		log.Printf("candidate=%s, candidate_last=%s", candidate, candidate_last)
		if bytes.Compare(*candidate_last, m.last) <= 0 {
			// We found one small enough (or exact sized) so use it
			block := BlockCreate(candidate, &m.value)
			log.Printf("buildAndEmit:   emit.net=%s, value='%s'", block.net.String(), *block.value)
			if err := m.Emit(block); err != nil {
				return err
			}
			// We've emitted part of our block, shift to the IP after what we just emitted
			m.first = *IpNext(candidate_last)
			log.Printf("buildAndEmit:   new first=%s", m.first.String())
		}
		// try one smaller
		bit = bit + 1
	}

	return nil
}

func (m *CombiningProcessor) Receive(id string, block *Block) error {
	log.Printf("Receive: block.net=%s, block.value=%s", block.net.String(), *block.value)
	if m.value != *block.value || bytes.Compare(m.next, block.net.IP) != 0 {
		// There's a change in value or gap in ips, we can't comebine
		if len(m.value) > 0 {
			// We have a non-empty current value so we need to emit whatever
			// we've built up
			if err := m.buildAndEmit(); err != nil {
				return err
			}
		}
		// A new starting point and value for current, last will get updated
		// below
		m.first = block.net.IP
		m.value = *block.value
	}

	// Always update last
	m.last = *NetLast(block.net)
	m.next = *IpNext(&m.last)

	return nil
}

func (m *CombiningProcessor) Done(id string) error {
	return m.Emitter.Done()
}
