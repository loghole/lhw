package batch

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	batch := NewBatch(0)

	t.Run("AppendBytes", func(t *testing.T) {
		msg := []byte("message")
		expected := []byte("[message]")

		batch.Reset()
		batch.AppendBytes(msg)

		assert.Equal(t, expected, batch.Bytes())
		assert.Equal(t, len(expected), batch.Len())
		assert.Equal(t, string(expected), batch.String())

		expected = []byte("[message,message]")

		batch.AppendBytes(msg)

		assert.Equal(t, expected, batch.Bytes())
		assert.Equal(t, len(expected), batch.Len())
		assert.Equal(t, string(expected), batch.String())
	})
}

func BenchmarkBatch_AppendBytes(b *testing.B) {
	str := bytes.Repeat([]byte("a"), 1024)

	slice := make([]byte, 1024)
	buf := bytes.NewBuffer(slice)
	batch := NewBatch(1024)

	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if buf.Len() > 0 {
				buf.WriteString(",")
			}

			buf.Write(str)
			buf.Reset()
		}
	})

	b.Run("CustomBatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			batch.AppendBytes(str)
			batch.Reset()
		}
	})
}
