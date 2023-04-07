package ssz

import (
	"encoding/binary"
)

type Encoder interface {
	MarshalSSZ() ([]byte, error)
	MarshalSSZTo(buf []byte) error
}

func EncodeBool(dst []byte, b bool) []byte {
	if b {
		return append(dst, byte(1))
	}
	return append(dst, byte(0))
}

func EncodeByte(dst []byte, i uint8) []byte {
	return append(dst, i)
}

func EncodeUint16(dst []byte, i uint16) []byte {
	dst = grow(dst, 2)
	return binary.LittleEndian.AppendUint16(dst, i)
}

func EncodeUint32(dst []byte, i uint32) []byte {
	dst = grow(dst, 4)
	return binary.LittleEndian.AppendUint32(dst, i)
}

func EncodeUint64(dst []byte, i uint64) []byte {
	dst = grow(dst, 8)
	return binary.LittleEndian.AppendUint64(dst, i)
}

func EncodeBools(dst []byte, input []bool) []byte {
	dst = grow(dst, len(input))
	for _, b := range input {
		if b {
			dst = append(dst, byte(1))
		}
		dst = append(dst, byte(0))
	}
	return dst
}

func EncodeBytes(dst []byte, b []byte) []byte {
	dst = grow(dst, len(b))
	return append(dst, b...)
}

func EncodeUint16s(dst []byte, input []uint16) []byte {
	dst = grow(dst, len(input)*2)
	for _, i := range input {
		dst = binary.LittleEndian.AppendUint16(dst, i)
	}
	return dst
}

func EncodeUint32s(dst []byte, input []uint32) []byte {
	dst = grow(dst, len(input)*4)
	for _, i := range input {
		dst = binary.LittleEndian.AppendUint32(dst, i)
	}
	return dst
}

func EncodeUint64s(dst []byte, input []uint64) []byte {
	dst = grow(dst, len(input)*8)
	for _, i := range input {
		dst = binary.LittleEndian.AppendUint64(dst, i)
	}
	return dst
}

func grow(buf []byte, n int) []byte {
	if cap(buf)-len(buf) < n {
		nbuf := make([]byte, len(buf), 2*cap(buf))
		copy(nbuf, buf)
		buf = nbuf
	}
	return buf
}
