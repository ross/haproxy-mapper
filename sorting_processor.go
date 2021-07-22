package main

import (
	"sort"
)

type SortingProcessor struct {
	sources  []Source
	prepared bool
	blocks   Blocks
}

func SortingProcessorCreate() *SortingProcessor {
	return &SortingProcessor{
		sources:  make([]Source, 0),
		prepared: false,
		blocks:   make(Blocks, 0),
	}
}

func (s *SortingProcessor) AddSource(source Source) {
	s.sources = append(s.sources, source)
}

func (s *SortingProcessor) prepare() error {
	s.prepared = true

	// TODO: if we make a rule that sources are sorted this can be simplified a
	// lot and avoid needing to load everything into memory
	for _, source := range s.sources {
		block, err := source.Next()
		for ; block != nil && err == nil; block, err = source.Next() {
			s.blocks = append(s.blocks, block)
		}
		if err != nil {
			return err
		}
	}

	sort.Sort(s.blocks)

	return nil
}

func (s *SortingProcessor) Next() (*Block, error) {
	if !s.prepared {
		if err := s.prepare(); err != nil {
			return nil, err
		}
	}

	n := len(s.blocks)
	if n > 0 {
		block := s.blocks[0]
		s.blocks = s.blocks[1:n]
		return block, nil
	}

	return nil, nil
}
