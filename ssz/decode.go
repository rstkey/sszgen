package ssz

import (
	"encoding/binary"
	"errors"
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
	b, err := s.readByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func DecodeUint16(s *Stream) (uint16, error) {
	buf, err := s.read(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf[:]), nil
}

func DecodeUint32(s *Stream) (uint32, error) {
	buf, err := s.read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func DecodeUint64(s *Stream) (uint64, error) {
	buf, err := s.read(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func DecodeBytes(s *Stream, n int) ([]byte, error) {
	return read(s, n)
}

func DecodeUint16s(s *Stream, n int) ([]uint16, error) {
	buf, err := read(s, n)
	if err != nil {
		return nil, err
	}
	if len(buf)%2 != 0 {
		return nil, errors.New("invalid input for decoding uint16s")
	}
	ret := make([]uint16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		ret[i] = binary.LittleEndian.Uint16(buf[2*i : 2*i+1])
	}
	return ret, nil
}

func DecodeUint32s(s *Stream, n int) ([]uint32, error) {
	buf, err := read(s, n)
	if err != nil {
		return nil, err
	}
	if len(buf)%4 != 0 {
		return nil, errors.New("invalid input for decoding uint32s")
	}
	ret := make([]uint32, len(buf)/4)
	for i := 0; i < len(buf)/4; i++ {
		ret[i] = binary.LittleEndian.Uint32(buf[4*i : 4*i+3])
	}
	return ret, nil
}

func DecodeUint64s(s *Stream, n int) ([]uint64, error) {
	buf, err := read(s, n)
	if err != nil {
		return nil, err
	}
	if len(buf)%8 != 0 {
		return nil, errors.New("invalid input for decoding uint64s")
	}
	ret := make([]uint64, len(buf)/8)
	for i := 0; i < len(buf)/8; i++ {
		ret[i] = binary.LittleEndian.Uint64(buf[8*i : 8*i+7])
	}
	return ret, nil
}

func read(s *Stream, n int) ([]byte, error) {
	if n == 0 {
		return s.readEnd()
	}
	return s.read(n)
}
