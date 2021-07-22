package main

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

func (s *SortingProcessor) Next() (*Block, error) {
	// TODO: actually sort
	block, err := s.current.Next()
	if block == nil && len(s.sources) > 0 {
		s.current = s.sources[0]
		s.sources = s.sources[1:len(s.sources)]
		block, err = s.current.Next()
	}

	return block, err
}
