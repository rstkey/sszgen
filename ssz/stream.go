package ssz

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

var (
	ErrValueTooLarge = errors.New("rlp: value size exceeds available input length")
)

type ByteReader interface {
	io.Reader
	io.ByteReader
}

type Stream struct {
	r         ByteReader
	remaining uint32 // number of bytes remaining to be read from r
	limited   bool   // true if input limit is in effect
}

func NewStream(r io.Reader, inputLimit uint32) *Stream {
	s := new(Stream)
	// Attempt to automatically discover
	// the limit when reading from a byte slice.
	switch br := r.(type) {
	case *bytes.Reader:
		s.remaining = uint32(br.Len())
		s.limited = true
	case *bytes.Buffer:
		s.remaining = uint32(br.Len())
		s.limited = true
	case *strings.Reader:
		s.remaining = uint32(br.Len())
		s.limited = true
	default:
		s.limited = false
	}
	// Wrap r with a buffer if it doesn't have one.
	bufr, ok := r.(ByteReader)
	if !ok {
		bufr = bufio.NewReader(r)
	}
	s.r = bufr

	s.Reset(inputLimit)
	return s
}

// Reset discards any information about the current decoding context
// and starts reading from r. This method is meant to facilitate reuse
// of a preallocated Stream across many decoding operations.
//
// If r does not also implement ByteReader, Stream will do its own
// buffering.
func (s *Stream) Reset(inputLimit uint32) {
	if inputLimit > 0 {
		s.remaining = inputLimit
		s.limited = true
	}
}

// readFull reads into buf from the underlying stream.
func (s *Stream) readFull(buf []byte) (err error) {
	if err := s.willRead(uint32(len(buf))); err != nil {
		return err
	}
	var nn, n int
	for n < len(buf) && err == nil {
		nn, err = s.r.Read(buf[n:])
		n += nn
	}
	if err == io.EOF {
		if n < len(buf) {
			err = io.ErrUnexpectedEOF
		} else {
			// Readers are allowed to give EOF even though the read succeeded.
			// In such cases, we discard the EOF, like io.ReadFull() does.
			err = nil
		}
	}
	return err
}

func (s *Stream) readAll() ([]byte, error) {
	buf := make([]byte, s.remaining)
	if err := s.readFull(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// readByte reads a single byte from the underlying stream.
func (s *Stream) readByte() (byte, error) {
	if err := s.willRead(1); err != nil {
		return 0, err
	}
	b, err := s.r.ReadByte()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return b, err
}

// willRead is called before any read from the underlying stream. It checks
// n against size limits, and updates the limits if n doesn't overflow them.
func (s *Stream) willRead(n uint32) error {
	if s.limited {
		if n > s.remaining {
			return ErrValueTooLarge
		}
		s.remaining -= n
	}
	return nil
}
