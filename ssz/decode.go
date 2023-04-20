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

func DecodeBool(buf []byte) ([]byte, bool, error) {
	if len(buf) < 1 {
		return nil, false, ErrX
	}
	return buf[1:], buf[0] == byte(1), nil
}

func DecodeByte(buf []byte) ([]byte, byte, error) {
	if len(buf) < 1 {
		return nil, 0, ErrX
	}
	return buf[1:], buf[0], nil
}

func DecodeUint16(buf []byte) ([]byte, uint16, error) {
	if len(buf) < 2 {
		return nil, 0, ErrX
	}
	return buf[2:], binary.LittleEndian.Uint16(buf[:2]), nil
}

func DecodeUint32(buf []byte) ([]byte, uint32, error) {
	if len(buf) < 4 {
		return nil, 0, ErrX
	}
	return buf[4:], binary.LittleEndian.Uint32(buf[:4]), nil
}

func DecodeUint64(buf []byte) ([]byte, uint64, error) {
	if len(buf) < 8 {
		return nil, 0, ErrX
	}
	return buf[8:], binary.LittleEndian.Uint64(buf[:8]), nil
}

func DecodeBytes(buf []byte) ([]byte, []byte, error) {
	return nil, buf, nil
}

func DecodeUint16s(buf []byte) ([]byte, []uint16, error) {
	if len(buf)%2 != 0 {
		return nil, nil, ErrX
	}
	var ret []uint16
	for i := 0; i < len(buf)/2; i++ {
		ret = append(ret, binary.LittleEndian.Uint16(buf[2*i:2*i+1]))
	}
	return nil, ret, nil
}

func DecodeUint32s(buf []byte) ([]byte, []uint32, error) {
	if len(buf)%4 != 0 {
		return nil, nil, ErrX
	}
	var ret []uint32
	for i := 0; i < len(buf)/4; i++ {
		ret = append(ret, binary.LittleEndian.Uint32(buf[4*i:4*i+3]))
	}
	return nil, ret, nil
}

func DecodeUint64s(buf []byte) ([]byte, []uint64, error) {
	if len(buf)%8 != 0 {
		return nil, nil, ErrX
	}
	var ret []uint64
	for i := 0; i < len(buf)/8; i++ {
		ret = append(ret, binary.LittleEndian.Uint64(buf[8*i:8*i+7]))
	}
	return nil, ret, nil
}
