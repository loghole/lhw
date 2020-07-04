package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Storage interface {
	Put(data []byte) error
	Pop() ([]byte, error)
	Drop() error
	IsUsed() bool
}

// New constructs a new logs storage.
// To create storage that does not persist to disk use :memory: as the path of the file.
func New(path string) (Storage, error) {
	if path == ":memory:" {
		return newMemoryStorage(), nil
	}

	return newFileStorage(path)
}

type MemoryStorage struct {
	mu      sync.Mutex
	storage [][]byte
}

func newMemoryStorage() *MemoryStorage {
	return &MemoryStorage{storage: make([][]byte, 0)}
}

func (s *MemoryStorage) Put(data []byte) error {
	s.mu.Lock()
	s.storage = append(s.storage, append(data[:0:0], data...))
	s.mu.Unlock()

	return nil
}

func (s *MemoryStorage) Pop() (b []byte, err error) {
	s.mu.Lock()

	if len(s.storage) == 0 {
		s.mu.Unlock()

		return nil, errors.New("no such data")
	}

	b, s.storage = s.storage[0], s.storage[1:]
	s.mu.Unlock()

	return b, nil
}

func (s *MemoryStorage) Drop() error {
	s.mu.Lock()
	s.storage = make([][]byte, 0)
	s.mu.Unlock()

	return nil
}

func (s *MemoryStorage) IsUsed() (ok bool) {
	s.mu.Lock()
	ok = len(s.storage) > 0
	s.mu.Unlock()

	return ok
}

type FileStorage struct {
	dir   string
	file  string
	count int64
}

func newFileStorage(filepath string) (*FileStorage, error) {
	s := new(FileStorage)
	s.dir, s.file = path.Split(filepath)

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *FileStorage) init() (err error) {
	if s.file == "" {
		s.file = "app.log"
	}

	err = os.MkdirAll(s.dir, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return err
	}

	s.count = int64(len(files))

	return nil
}

func (s *FileStorage) Put(data []byte) (err error) {
	err = ioutil.WriteFile(s.filename(), data, os.ModePerm)
	if err != nil {
		return err
	}

	atomic.AddInt64(&s.count, 1)

	return nil
}

func (s *FileStorage) Pop() (data []byte, err error) {
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		atomic.StoreInt64(&s.count, 0)

		return nil, errors.New("no such data")
	}

	filename := fmt.Sprint(s.dir, files[0].Name())

	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = os.Remove(filename); err != nil {
		return nil, err
	}

	atomic.AddInt64(&s.count, -1)

	return data, nil
}

func (s *FileStorage) Drop() (err error) {
	return os.RemoveAll(s.dir)
}

func (s *FileStorage) IsUsed() bool {
	return atomic.LoadInt64(&s.count) > 0
}

func (s *FileStorage) filename() string {
	t := strconv.FormatInt(time.Now().UnixNano(), 10)

	buf := make([]byte, 0, len(s.dir)+len(t)+len(s.file)+1)
	buf = append(buf, s.dir...)
	buf = append(buf, s.file...)
	buf = append(buf, "."...)
	buf = append(buf, t...)

	return string(buf)
}
