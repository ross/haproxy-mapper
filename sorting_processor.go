package main

import (
	"net"
)

type SortingProcessor struct {
	sources []Source
	current Source
}

func SortingProcessorCreate() *SortingProcessor {
	return &SortingProcessor{
		sources: make([]Source, 0),
		current: nil,
	}
}

func (s *SortingProcessor) AddSource(source Source) {
	if s.current == nil {
		s.current = source
	} else {
		s.sources = append(s.sources, source)
	}
}

func (s *SortingProcessor) Next() (*net.IPNet, *string, error) {
	// TODO: actually sort
	net, value, err := s.current.Next()
	if (net == nil || value == nil) && len(s.sources) > 0 {
		s.current = s.sources[0]
		s.sources = s.sources[1:len(s.sources)]
		net, value, err = s.current.Next()
	}

	return net, value, err
}
