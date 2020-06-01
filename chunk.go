package main

import "github.com/pkg/errors"

const (
	TimestampThreshold = 16777215
)

var (
	FmtInvalidError = errors.New("Fmt is invalid.")
)

type Chunk struct {
	Header *ChunkHeader
	Data []byte
}

type ChunkHeader struct {
	BasicHeader *BasicHeader
	MessageHeader *MessageHeader
	ExtendedTimestamp uint32
}

type BasicHeader struct {
	Fmt uint8
	Csid uint32
}

type MessageHeader struct {
	Timestamp uint32
	TimestampDelta uint32
	MessageLength uint32
	MessageTypeID uint8
	MessageStreamID uint32
}

func ReadBasicHeader(proxy *NetworkProxy) (*BasicHeader, error) {
	header := &BasicHeader{}
	u, e := proxy.ReadByte()
	if e != nil {
		return nil, e
	}

	header.Fmt = u >> 6
	csid := u & 0x3F
	if csid == 0 {
		csid_64, err := proxy.ReadByte()
		if err != nil {
			return nil, err
		}
		header.Csid = uint32(csid_64 + 64)
	} else if csid == 1 {
		csid_64, err := proxy.ReadUint32(2)
		if err != nil {
			return nil, err
		}
		header.Csid = csid_64 + 64
	} else {
		header.Csid = uint32(csid)
	}

	return header, nil
}

func ReadMessageHeader(fmt uint8, proxy *NetworkProxy) (*MessageHeader, error) {
	mh := &MessageHeader{}

	switch fmt {
	case 0:
		timestamp, err := proxy.ReadUint32(3)
		if err != nil {
			return nil, err
		}
		mh.Timestamp = timestamp

		length, err := proxy.ReadUint32(3)
		if err != nil {
			return nil, err
		}
		mh.MessageLength = length

		typeID, err := proxy.ReadByte()
		if err != nil {
			return nil, err
		}
		mh.MessageTypeID = typeID

		streamID, err := proxy.ReadUint32()
		if err != nil {
			return nil, err
		}
		mh.MessageStreamID = streamID
		return mh, nil
	case 1:
		timestampDelta, err := proxy.ReadUint32(3)
		if err != nil {
			return nil, err
		}
		mh.TimestampDelta = timestampDelta

		length, err := proxy.ReadUint32(3)
		if err != nil {
			return nil, err
		}
		mh.MessageLength = length

		typeID, err := proxy.ReadByte()
		if err != nil {
			return nil, err
		}
		mh.MessageTypeID = typeID
		return mh, nil
	case 2:
		timestampDelta, err := proxy.ReadUint32(3)
		if err != nil {
			return nil, err
		}
		mh.TimestampDelta = timestampDelta
		return mh, nil
	case 3:
		return mh, nil
	default:
		return nil, FmtInvalidError
	}
}

func ReadChunkHeader(proxy *NetworkProxy) (*ChunkHeader, error) {
	ch := &ChunkHeader{
		BasicHeader:       nil,
		MessageHeader:     nil,
		ExtendedTimestamp: 0,
	}

	basicHeader, e := ReadBasicHeader(proxy)
	if e != nil {
		return nil, e
	}
	ch.BasicHeader = basicHeader

	messageHeader, e := ReadMessageHeader(basicHeader.Fmt, proxy)
	if e != nil {
		return nil, e
	}
	ch.MessageHeader = messageHeader

	if ch.MessageHeader.Timestamp >= TimestampThreshold {
		et, err := proxy.ReadUint32()
		if err != nil {
			return nil, err
		}
		ch.ExtendedTimestamp = et
	} else if ch.MessageHeader.TimestampDelta >= TimestampThreshold {
		et, err := proxy.ReadUint32()
		if err != nil {
			return nil, err
		}
		ch.ExtendedTimestamp = et
	}

	return ch, nil
}
