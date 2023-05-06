package ssz

import (
	"encoding/binary"
	"errors"
)

var (
	ErrX = errors.New("xx")
)

type Decoder interface {
	UnmarshalSSZ(buf []byte) error
}

func DecodeBool(s *Stream) (bool, error) {
	b, err := s.readByte()
	if err != nil {
		return false, err
	}
	return b == byte(1), nil
}

func DecodeByte(s *Stream) (byte, error) {
	return s.readByte()
}

func DecodeUint16(s *Stream) (uint16, error) {
	var buf [2]byte
	if err := s.readFull(buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf[:2]), nil
}

func DecodeUint32(s *Stream) (uint32, error) {
	var buf [4]byte
	if err := s.readFull(buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func DecodeUint64(s *Stream) (uint64, error) {
	var buf [8]byte
	if err := s.readFull(buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:8]), nil
}

func DecodeBytes(s *Stream) ([]byte, error) {
	return s.readAll()
}

func DecodeUint16s(s *Stream) ([]uint16, error) {
	buf, err := s.readAll()
	if err != nil {
		return nil, err
	}
	if len(buf)%2 != 0 {
		return nil, ErrX
	}
	var ret []uint16
	for i := 0; i < len(buf)/2; i++ {
		ret = append(ret, binary.LittleEndian.Uint16(buf[2*i:2*i+1]))
	}
	return ret, nil
}

func DecodeUint32s(s *Stream) ([]uint32, error) {
	buf, err := s.readAll()
	if err != nil {
		return nil, err
	}
	if len(buf)%4 != 0 {
		return nil, ErrX
	}
	var ret []uint32
	for i := 0; i < len(buf)/4; i++ {
		ret = append(ret, binary.LittleEndian.Uint32(buf[4*i:4*i+3]))
	}
	return ret, nil
}

func DecodeUint64s(s *Stream) ([]uint64, error) {
	buf, err := s.readAll()
	if err != nil {
		return nil, err
	}
	if len(buf)%8 != 0 {
		return nil, ErrX
	}
	var ret []uint64
	for i := 0; i < len(buf)/8; i++ {
		ret = append(ret, binary.LittleEndian.Uint64(buf[8*i:8*i+7]))
	}
	return ret, nil
}
