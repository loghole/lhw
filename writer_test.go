package lhw

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gadavy/lhw/test"
)

func TestWriter_Write(t *testing.T) {
	tests := []struct {
		name        string
		writer      func() *Writer
		wantErr     bool
		expectedN   int
		expectedErr string
	}{
		{
			name: "Pass",
			writer: func() *Writer {
				queue := make(chan []byte, 1)

				return &Writer{queue: queue}
			},
			wantErr:     false,
			expectedN:   12,
			expectedErr: "",
		},
		{
			name: "ErrorQueueIsFull",
			writer: func() *Writer {
				queue := make(chan []byte, 1)
				queue <- []byte{}

				return &Writer{queue: queue}
			},
			wantErr:     true,
			expectedN:   0,
			expectedErr: "write data to queue failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := tt.writer()

			n, err := writer.Write([]byte("test message"))
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.Equal(t, tt.expectedN, n)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestWriterQueue(t *testing.T) {
	transport := &test.StubTransport{}

	writer := &Writer{
		transport: transport,
		queue:     make(chan []byte, 1000),
		closed:    make(chan struct{}, 1),
	}

	for i := 0; i < 5; i++ {
		go func() {
			for i := 0; i < 100; i++ {
				writer.Write([]byte{})
			}
		}()
	}

	writer.wg.Add(1)
	go writer.worker()

	writer.Close()

	assert.Equal(t, int64(500), transport.Counter)
}
