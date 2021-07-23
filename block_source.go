package main

import (
	"sort"
)

type Loadable interface {
	Load(bool) (Blocks, error)
}

type BlockSource struct {
	Ipv4Only bool
	loadable Loadable
	blocks   Blocks
	loaded   bool
}

func BlockSourceCreate(loadable Loadable, ipv4Only bool) *BlockSource {
	return &BlockSource{
		Ipv4Only: ipv4Only,
		loadable: loadable,
		blocks:   make(Blocks, 0),
		loaded:   false,
	}
}

func (s *BlockSource) Next() (*Block, error) {
	if !s.loaded {
		s.loaded = true

		blocks, err := s.loadable.Load(s.Ipv4Only)
		if err != nil {
			return nil, err
		}

		sort.Sort(blocks)
		s.blocks = blocks
	}

	n := len(s.blocks)
	if n > 0 {
		block := s.blocks[0]
		s.blocks = s.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
