package ssz

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrValueTooLarge = errors.New("ssz: value size exceeds available input length")
)

type ByteReader interface {
	io.Reader
	io.ByteReader
}

type Stream struct {
	reader    ByteReader
	remaining uint32

	// [4, 10, 14]
	stack []uint32

	// [END - lastOffset]
	lastOffset uint32
	current    uint32
}

func NewStream(r io.Reader, size uint32) (*Stream, error) {
	var remaining uint32
	switch br := r.(type) {
	case *bytes.Reader:
		remaining = uint32(br.Len())
	case *bytes.Buffer:
		remaining = uint32(br.Len())
	case *strings.Reader:
		remaining = uint32(br.Len())
	}
	if size != 0 {
		if remaining != 0 && remaining != size {
			return nil, fmt.Errorf("invalid stream size, has: %d, want: %d", remaining, size)
		}
		remaining = size
	}
	// Wrap r with a buffer if it doesn't have one.
	bufr, ok := r.(ByteReader)
	if !ok {
		bufr = bufio.NewReader(r)
	}
	return &Stream{
		reader:    bufr,
		remaining: remaining,
	}, nil
}

// read reads into buf from the underlying stream.
func (s *Stream) read(n int) ([]byte, error) {
	if err := s.willRead(uint32(n)); err != nil {
		return nil, err
	}
	var (
		read int
		nn   int
		err  error
		buf  = make([]byte, n)
	)
	for read < n && err == nil {
		nn, err = s.reader.Read(buf[read:])
		read += nn
	}
	if err == io.EOF {
		if read < n {
			err = io.ErrUnexpectedEOF
		} else {
			// Readers are allowed to give EOF even though the read succeeded.
			// In such cases, we discard the EOF, like io.ReadFull() does.
			err = nil
		}
	}
	return buf, err
}

// readByte reads a single byte from the underlying stream.
func (s *Stream) readByte() (byte, error) {
	if err := s.willRead(1); err != nil {
		return 0, err
	}
	b, err := s.reader.ReadByte()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return b, err
}

func (s *Stream) readEnd() ([]byte, error) {
	if s.current != 0 {
		return s.read(int(s.current))
	}
	if len(s.stack) != 0 {
		return nil, errors.New("unexpected readEnd operation")
	}
	if s.remaining != 0 {
		return s.read(int(s.remaining))
	}
	var (
		nn, n    int
		err      error
		buf      []byte
		internal = make([]byte, 1024)
	)
	for err == nil {
		nn, err = s.reader.Read(internal)
		n += nn
		buf = append(buf, internal[:nn]...)
	}
	// Readers are allowed to give EOF even though the read succeeded.
	// In such cases, we discard the EOF, like io.ReadFull() does.
	if err == io.EOF {
		err = nil
	}
	return buf, err
}

// willRead is called before any read from the underlying stream. It checks
// n against size limits, and updates the limits if n doesn't overflow them.
func (s *Stream) willRead(n uint32) error {
	if s.remaining > 0 {
		if n > s.remaining {
			return ErrValueTooLarge
		}
		s.remaining -= n
	}
	if s.current != 0 {
		if n > s.current {
			return ErrValueTooLarge
		}
		s.current -= n
	}
	return nil
}

func (s *Stream) decodeOffset() (uint32, error) {
	buf, err := s.read(4)
	if err != nil {
		return 0, err
	}
	offset := binary.LittleEndian.Uint32(buf)
	if s.lastOffset == 0 {
		s.lastOffset = offset
		return offset, nil
	}
	s.stack = append(s.stack, offset-s.lastOffset)
	s.lastOffset = offset
	return offset, nil
}

func (s *Stream) DecodeOffset() error {
	_, err := s.decodeOffset()
	return err
}

func (s *Stream) ReadOffset() (uint32, error) {
	return s.decodeOffset()
}

func (s *Stream) BlockStart() error {
	if s.current != 0 {
		return errors.New("last block is not fully consumed")
	}
	if len(s.stack) == 0 {
		return errors.New("no block available")
	}
	// TODO
	return nil
}

func (s *Stream) BlockEnd() error {
	// TODO
	return nil
}
