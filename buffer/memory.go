package buffer

import (
	"io"

	"github.com/pkg/errors"
)

type Memory struct {
	data []byte
	pos  int64
}

func NewMemoryBuffer() *Memory {
	return &Memory{}
}

func (m *Memory) Len() int {
	if m.pos >= int64(len(m.data)) {
		return 0
	}
	return int(int64(len(m.data)) - m.pos)
}

func (m *Memory) Size() int64 {
	return int64(len(m.data))
}

func (m *Memory) Read(b []byte) (n int, err error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(b, m.data[m.pos:])
	m.pos += int64(n)
	return n, err
}

func (m *Memory) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = m.pos + offset
	case io.SeekEnd:
		abs = int64(len(m.data)) + offset
	default:
		return 0, errors.New("Memory: invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("Memory: negative position")
	}
	m.pos = abs
	return abs, nil
}

func (m *Memory) Close() error {
	m.data = nil
	m.pos = 0
	return nil
}

func (m *Memory) Pos() int64 {
	return m.pos
}

func (m *Memory) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	m.pos += int64(len(m.data))
	return len(p), nil
}

func (m *Memory) Bytes() []byte {
	return m.data
}
