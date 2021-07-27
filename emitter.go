package main

type Receiver interface {
	Subscribed(id string)
	Receive(id string, block *Block) error
	Done(id string) error
}

type Emitter struct {
	id        string
	receivers []Receiver
}

func EmitterCreate() Emitter {
	return Emitter{
		receivers: make([]Receiver, 0),
	}
}

func (e *Emitter) AddReceiver(receiver Receiver) {
	e.receivers = append(e.receivers, receiver)
	receiver.Subscribed(e.id)
}

func (e *Emitter) Emit(block *Block) error {
	for _, receiver := range e.receivers {
		err := receiver.Receive(e.id, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Emitter) Done() error {
	for _, receiver := range e.receivers {
		err := receiver.Done(e.id)
		if err != nil {
			return err
		}
	}
	return nil
}