package main

import (
	"bufio"
	"errors"
	"os"
)

type Map struct {
	Filename string
	fh       *os.File
	buf      *bufio.Writer
}

func MapCreate(filename string) (*Map, error) {
	fh, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &Map{
		Filename: filename,
		fh:       fh,
		buf:      bufio.NewWriter(fh),
	}, nil
}

func (m *Map) Close() error {
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

func (m *Map) Write(block *Block) error {
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

func (m *Map) Consume(source Source) error {
	block, err := source.Next()
	for ; block != nil && err == nil; block, err = source.Next() {
		if len(*block.value) > 0 {
			err = m.Write(block)
		}
	}
	return err
}
