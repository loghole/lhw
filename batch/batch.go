package batch

import (
	"bytes"
)

const (
	comma = ','
)

type Batch struct {
	buf []byte
}

func NewBatch(size int) *Batch {
	return &Batch{buf: make([]byte, 0, size)}
}

func (b *Batch) AppendBytes(data []byte) {
	if len(b.buf) > 0 {
		b.appendComma()
	}

	b.buf = append(b.buf, data...)
}

func (b *Batch) Bytes() []byte {
	return bytes.Join([][]byte{[]byte("["), b.buf[0:], []byte("]")}, []byte{})
}

func (b *Batch) String() string {
	return string(b.Bytes())
}

func (b *Batch) Len() int {
	return len(b.buf) + 2
}

func (b *Batch) Reset() {
	b.buf = b.buf[:0]
}

func (b *Batch) appendComma() {
	if b.buf[len(b.buf)-1] != comma {
		b.buf = append(b.buf, comma)
	}
}
