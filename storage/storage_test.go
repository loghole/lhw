package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Memory", func(t *testing.T) {
		storage, err := New(":memory:")
		if err != nil {
			t.Error(err)
		}

		assert.IsType(t, &MemoryStorage{}, storage)
	})

	t.Run("File", func(t *testing.T) {
		storage, err := New("logs/tmp.log")
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &FileStorage{}, storage)
		assert.Equal(t, "logs/", storage.(*FileStorage).dir)
		assert.Equal(t, "tmp.log", storage.(*FileStorage).file)

		if err = storage.Drop(); err != nil {
			t.Error(err)
		}

		storage, err = New("logs/")
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &FileStorage{}, storage)
		assert.Equal(t, "logs/", storage.(*FileStorage).dir)
		assert.Equal(t, "app.log", storage.(*FileStorage).file)

		if err = storage.Drop(); err != nil {
			t.Error(err)
		}

		_, err = New("")

		assert.EqualError(t, err, "mkdir : no such file or directory")

		if err = storage.Drop(); err != nil {
			t.Error(err)
		}
	})
}

func TestFileStorage(t *testing.T) {
	storage, err := newFileStorage("logs/tmp.log")
	if err != nil {
		t.Fatal(err)
	}

	var (
		msg1 = []byte("message 1")
		msg2 = []byte("message 2")
		msg3 = []byte("message 3")
	)

	err = storage.Put(msg1)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Put(msg2)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Put(msg3)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, (int64)(3), storage.count)
	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err := storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg1, msg)
	assert.Equal(t, (int64)(2), storage.count)
	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err = storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg2, msg)
	assert.Equal(t, (int64)(1), storage.count)
	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err = storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg3, msg)
	assert.Equal(t, (int64)(0), storage.count)
	assert.False(t, storage.IsUsed(), "expected is not used")

	_, err = storage.Pop()
	assert.EqualError(t, err, "no such data")
	assert.False(t, storage.IsUsed(), "expected is not used")

	if err = storage.Drop(); err != nil {
		t.Error(err)
	}

	err = os.MkdirAll("logs/tmp.log", os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("logs/tmp.log.1", []byte("message"), os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	storage, err = newFileStorage("logs/tmp.log")
	if err != nil {
		t.Error(err)
	}

	assert.True(t, storage.IsUsed(), "expected is used")

	if err = storage.Drop(); err != nil {
		t.Error(err)
	}
}

func TestMemoryStorage(t *testing.T) {
	storage := newMemoryStorage()

	var (
		msg1 = []byte("message 1")
		msg2 = []byte("message 2")
		msg3 = []byte("message 3")
	)

	err := storage.Put(msg1)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Put(msg2)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Put(msg3)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err := storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg1, msg)
	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err = storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg2, msg)
	assert.True(t, storage.IsUsed(), "expected is used")

	msg, err = storage.Pop()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg3, msg)
	assert.False(t, storage.IsUsed(), "expected is not used")

	_, err = storage.Pop()
	assert.EqualError(t, err, "no such data")
	assert.False(t, storage.IsUsed(), "expected is not used")

	if err = storage.Drop(); err != nil {
		t.Fatal(err)
	}
}
