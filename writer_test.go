package lhw

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gadavy/lhw/internal"
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
				return &Writer{queue: internal.NewQueue(1)}
			},
			wantErr:     false,
			expectedN:   12,
			expectedErr: "",
		},
		{
			name: "ErrorQueueIsFull",
			writer: func() *Writer {
				queue := internal.NewQueue(0)
				_ = queue.Push([]byte("start msg"))

				return &Writer{queue: queue}
			},
			wantErr:     true,
			expectedN:   0,
			expectedErr: "write data to queue failed: queue is full",
		},
		{
			name: "ErrorQueueIsClosed",
			writer: func() *Writer {
				queue := internal.NewQueue(0)
				queue.Close()

				return &Writer{queue: queue}
			},
			wantErr:     true,
			expectedN:   0,
			expectedErr: "write data to queue failed: queue is closed",
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
