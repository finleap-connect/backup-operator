package util

import (
	"io"
	"io/ioutil"
)

func NewBufferDestination() (*BufferDestination, error) {
	return &BufferDestination{}, nil
}

type BufferDestination struct {
	Data []byte
}

func (b *BufferDestination) Store(data io.Reader) error {
	var err error
	b.Data, err = ioutil.ReadAll(data)
	return err
}
