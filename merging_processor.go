package main

type MergingProcessor struct {
	queues map[string][]*Block
	done   map[string]bool
	Emitter
}

func MergingProcessorCreate(id string) *MergingProcessor {
	return &MergingProcessor{
		queues: make(map[string][]*Block, 0),
		done:   make(map[string]bool, 0),
		Emitter: Emitter{
			id: id,
		},
	}
}

func (m *MergingProcessor) Subscribed(id string) {
	m.queues[id] = make([]*Block, 0)
}

func (m *MergingProcessor) Header(id, header string) error {
	// TODO: header specific to the merging processor
	for _, receiver := range m.receivers {
		if err := receiver.Header(id, header); err != nil {
			return err
		}
	}

	return nil
}

func (m *MergingProcessor) Receive(id string, block *Block) error {
	m.queues[id] = append(m.queues[id], block)
	return m.checkEmit()
}

func (m *MergingProcessor) checkEmit() error {
	// This will run until it finds an empty queue that's not done or fails to find anything
	for {
		var min *Block
		var minId string
		for id, queue := range m.queues {
			if len(queue) == 0 {
				if m.done[id] {
					// It's empty and done so that's Ok. We should get rid of
					// it and keep checking others
					delete(m.queues, id)
					continue
				}
				// We have an empty queue and therefore can't know what to emit
				// next, maybe next time through
				return nil
			}

			// If we don't yet have a min or this one is less than what we have
			if min == nil || queue[0].Less(min) {
				min = queue[0]
				minId = id
			}
		}
		if min == nil {
			return nil
		}

		// Pop the first item off
		m.queues[minId] = m.queues[minId][1:len(m.queues[minId])]

		// Emit it (min)
		err := m.Emit(min)
		if err != nil {
			return err
		}
	}
}

func (m *MergingProcessor) Done(id string) error {
	m.done[id] = true

	err := m.checkEmit()
	if err != nil {
		return err
	}

	if len(m.queues) == 0 {
		return m.Emitter.Done()
	}

	return nil
}
