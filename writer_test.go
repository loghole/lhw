package lhw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter_Write(t *testing.T) {
	tests := []struct {
		name        string
		writer      *Writer
		wantErr     bool
		expectedN   int
		expectedErr string
	}{
		{
			name:        "Pass",
			writer:      &Writer{queue: make(chan []byte, 1)},
			wantErr:     false,
			expectedN:   12,
			expectedErr: "",
		},
		{
			name:        "Error",
			writer:      &Writer{},
			wantErr:     true,
			expectedN:   0,
			expectedErr: ErrWriteFailed.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.writer.Write([]byte("test message"))
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
