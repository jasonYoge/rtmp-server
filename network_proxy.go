package main

import (
	"bufio"
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"net"
)

var (
	LimitOverflowError = errors.New("Beyond read size limit.")
	DataMissingError = errors.New("Write data size is fewer than expected.")
)

type NetworkProxy struct {
	conn net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewNetworkProxy(conn net.Conn) *NetworkProxy {
	return &NetworkProxy{
		conn:conn,
		reader:bufio.NewReader(conn),
		writer:bufio.NewWriter(conn),
	}
}

func (r *NetworkProxy) ReadByte() (uint8, error) {
	b, e := r.reader.ReadByte()
	if e != nil {
		return 0, e
	}

	return b, nil
}

func (r *NetworkProxy) ReadUint32(param ...uint) (uint32, error) {
	// default uint32 byte size
	var readSize uint = 4
	if len(param) > 0 {
		readSize = param[0]
	}

	if readSize > 4{
		return 0, LimitOverflowError
	}

	data := make([]byte, 4)
	_, err := r.reader.Read(data[4-readSize:])
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(data), nil
}

func (r *NetworkProxy) ReadNBytes(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := io.ReadFull(r.reader, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *NetworkProxy) WriteByte(data uint8) error {
	err := r.writer.WriteByte(data)
	if err != nil {
		return err
	}

	return nil
}

func (r *NetworkProxy) WriteBytes(data []byte) error {
	nn, err := r.writer.Write(data)
	if err != nil {
		return err
	}

	if nn < len(data) {
		return DataMissingError
	}

	return nil
}