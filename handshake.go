package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

const (
	ChunkSize = 1536
)

var (
	C0VersionInvalidError = errors.New("C0 version is invalid.")
	TimestampNotMatchError = errors.New("Timestamp not match.")
)

type HandshakeChunk interface {
	Convert2Bytes() []byte
}

type C1S1Chunk struct {
	Time uint32
	Zero uint32
	RandomBytes []byte
}

func (c *C1S1Chunk) Convert2Bytes() []byte {
	chunk := make([]byte, ChunkSize)
	binary.BigEndian.PutUint32(chunk[:4], c.Time)
	binary.BigEndian.PutUint32(chunk[4:8], c.Zero)
	copy(chunk[9:], c.RandomBytes)

	return chunk
}

type C2S2Content struct {
	Time uint32
	Time2 uint32
	RandomEcho []byte
}

func (c *C2S2Content) Convert2Bytes() []byte {
	chunk := make([]byte, ChunkSize)
	binary.BigEndian.PutUint32(chunk[:4], c.Time)
	binary.BigEndian.PutUint32(chunk[4:8], c.Time2)
	copy(chunk[9:], c.RandomEcho)

	return chunk
}

func handleC0S0(proxy *NetworkProxy) error {
	fmt.Println("Version sent start...")
	c0, err := proxy.ReadByte()
	if err != nil {
		return err
	}

	if c0 > 32 {
		return C0VersionInvalidError
	}

	// version s0 == c0
	s0 := c0
	err = proxy.WriteByte(s0)
	fmt.Println("Version sent end...")
	if err != nil {
		return err
	}

	proxy.writer.Flush()
	return nil
}

func handleC1S1(proxy *NetworkProxy) (c1 *C1S1Chunk, s1 *C1S1Chunk, err error) {
	fmt.Println("Ack sent start...")
	t, err := proxy.ReadUint32()
	if err != nil {
		return
	}

	z, err := proxy.ReadUint32()
	if err != nil {
		return
	}

	bytes, err := proxy.ReadNBytes(1528)
	if err != nil {
		return
	}

	c1 = &C1S1Chunk{
		Time:        t,
		Zero:        z,
		RandomBytes: bytes,
	}

	randBytes := make([]byte, 1528)
	rand.Read(randBytes)
	s1 = &C1S1Chunk{
		Time:        uint32(time.Now().Unix()),
		Zero:        0,
		RandomBytes: randBytes,
	}

	err = proxy.WriteBytes(s1.Convert2Bytes())

	proxy.writer.Flush()
	fmt.Println("Ack sent end...")
	return
}

func handleC2S2(proxy *NetworkProxy, c1 *C1S1Chunk, s1 *C1S1Chunk) error {
	fmt.Println("Final handshake start...")
	t1, err := proxy.ReadUint32()
	if err != nil {
		return err
	}
	if t1 != s1.Time {
		return TimestampNotMatchError
	}

	_, _ = proxy.ReadUint32()

	_, err = proxy.ReadNBytes(1528)
	if err != nil {
		return err
	}
	
	s2 := &C2S2Content{
		Time:       c1.Time,
		Time2:      s1.Time,
		RandomEcho: c1.RandomBytes,
	}

	err = proxy.WriteBytes(s2.Convert2Bytes())
	if err != nil {
		return err
	}

	proxy.writer.Flush()
	fmt.Println("Handshake done...")
	return nil
}

func Handshake(proxy *NetworkProxy) error {
	err := handleC0S0(proxy)
	if err != nil {
		return err
	}

	c1, s1, err := handleC1S1(proxy)
	if err != nil {
		return err
	}

	err = handleC2S2(proxy, c1, s1)
	if err != nil {
		return err
	}

	return nil
}