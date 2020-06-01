package main

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/pkg/errors"
	"log"
	"time"
)

const (
	ChunkSize = 1536
)

var (
	C0VersionInvalidError = errors.New("C0 version is invalid.")
	TimestampNotMatchError = errors.New("Timestamp not match.")
)

type ChunkType uint8

const (
	C0S0ChunkType ChunkType = iota
	C1S1ChunkType
	C2S2ChunkType
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

type C2S2Chunk struct {
	Time uint32
	Time2 uint32
	RandomEcho []byte
}

func (c *C2S2Chunk) Convert2Bytes() []byte {
	chunk := make([]byte, ChunkSize)
	binary.BigEndian.PutUint32(chunk[:4], c.Time)
	binary.BigEndian.PutUint32(chunk[4:8], c.Time2)
	copy(chunk[9:], c.RandomEcho)

	return chunk
}

func NewChunkFactory(cat ChunkType, arg1 uint32, arg2 uint32, random []byte) HandshakeChunk {
	switch cat {
	case C1S1ChunkType:
		return &C1S1Chunk{
			Time:        arg1,
			Zero:        arg2,
			RandomBytes: random,
		}
	case C2S2ChunkType:
		return &C2S2Chunk{
			Time:       arg1,
			Time2:      arg2,
			RandomEcho: random,
		}
	default:
		return nil
	}
}

func handleC0S0(proxy *NetworkProxy) error {
	log.Println("Version receiving starts...")
	c0, err := proxy.ReadByte()
	if err != nil {
		return err
	}

	if c0 > 32 {
		return C0VersionInvalidError
	}

	// version s0 == c0
	log.Println("Version sending ends...")
	s0 := c0
	err = proxy.WriteByte(s0)
	if err != nil {
		return err
	}

	return nil
}

func handleC1S1(proxy *NetworkProxy) (c1 *C1S1Chunk, s1 *C1S1Chunk, err error) {
	log.Println("Ack receiving starts...")
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

	c1 = NewChunkFactory(C1S1ChunkType, t, z, bytes).(*C1S1Chunk)

	randBytes := make([]byte, 1528)
	_, _ = rand.Read(randBytes)

	s1 = NewChunkFactory(C1S1ChunkType, uint32(time.Now().Unix()), 0, randBytes).(*C1S1Chunk)
	err = proxy.WriteBytes(s1.Convert2Bytes())

	log.Println("Ack sending ends...")
	return
}

func handleC2S2(proxy *NetworkProxy, c1 *C1S1Chunk, s1 *C1S1Chunk) error {
	log.Println("Final handshake phase starts...")
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

	s2 := NewChunkFactory(C2S2ChunkType, c1.Time, s1.Time, c1.RandomBytes).(*C2S2Chunk)

	err = proxy.WriteBytes(s2.Convert2Bytes())
	if err != nil {
		return err
	}

	log.Println("Handshake phase done...")
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

	err = proxy.writer.Flush()
	if err != nil {
		return err
	}

	err = handleC2S2(proxy, c1, s1)
	if err != nil {
		return err
	}

	err = proxy.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}