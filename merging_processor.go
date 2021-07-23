package main

type sourceAndCurrent struct {
	source  Source
	current *Block
}

type MergingProcessor struct {
	sourceAndCurrents []*sourceAndCurrent
}

func MergingProcessorCreate() *MergingProcessor {
	return &MergingProcessor{
		sourceAndCurrents: make([]*sourceAndCurrent, 0),
	}
}

func (s *MergingProcessor) AddSource(source Source) {
	s.sourceAndCurrents = append(s.sourceAndCurrents, &sourceAndCurrent{
		source:  source,
		current: nil,
	})
}

func (s *MergingProcessor) Next() (*Block, error) {
	if len(s.sourceAndCurrents) == 0 {
		return nil, nil
	}

	var err error
	var min *Block
	var minI int
	for i, sac := range s.sourceAndCurrents {
		// Make sure the sac has a current loaded
		if sac.current == nil {
			// It should be safe to call next on an empty source, they should
			// continue to return nil, nil, and sac's current will just stay nil
			sac.current, err = sac.source.Next()
			if err != nil {
				return nil, err
			}
		}

		if min == nil {
			// We don't yet have a min, this one should be it. If it's nil
			// that'll be Ok, either something will beat it or it'll win and
			// return nil below
			min = sac.current
			minI = i
		} else if sac.current != nil {
			// We have a min and a current so see which is less
			if sac.current.Less(min) {
				// We have a new min
				min = sac.current
				minI = i
			}
		}
	}

	// We have something to return, clear it from its sac
	s.sourceAndCurrents[minI].current = nil
	return min, nil
}
