package main

import (
	"bufio"
	"errors"
	"os"
)

type MapDestination struct {
	Filename string
	fh       *os.File
	buf      *bufio.Writer
}

func MapDestinationCreate(filename string) (*MapDestination, error) {
	fh, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &MapDestination{
		Filename: filename,
		fh:       fh,
		buf:      bufio.NewWriter(fh),
	}, nil
}

func (m *MapDestination) Subscribed(id string) {}

func (m *MapDestination) Header(id string, header string) error {
	var err error
	if len(header) > 0 {
		_, err = m.buf.WriteString(header)
	}
	return err
}

func (m *MapDestination) Receive(id string, block *Block) error {
	if m.buf == nil {
		return errors.New("Write called on closed Map")
	}

	if _, err := m.buf.WriteString(block.net.String()); err != nil {
		return err
	}
	if err := m.buf.WriteByte(' '); err != nil {
		return err
	}
	if _, err := m.buf.WriteString(*block.value); err != nil {
		return err
	}
	return m.buf.WriteByte('\n')
}

func (m *MapDestination) Done(id string) error {
	defer func() {
		m.buf = nil
		m.fh = nil
	}()

	var err error = nil
	if m.buf != nil {
		err = m.buf.Flush()
	}
	if m.fh != nil {
		return m.fh.Close()
	}
	return err
}
