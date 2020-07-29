package lhw

import (
	"errors"
	"strings"
	"sync"

	"github.com/gadavy/lhw/transport"
)

var (
	ErrWriteFailed = errors.New("write data to queue failed")
)

// The url can contain secret token e.g. https://secret_token@localhost:50000
// Comma separated arrays are also supported, e.g. urlA, urlB.
// Options start with the defaults but can be overridden.
func NewWriter(url string, options ...Option) (writer *Writer, err error) {
	opts := GetDefaultOptions()
	opts.Servers = processUrlString(url)

	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option(opts); err != nil {
			return nil, err
		}
	}

	writer = &Writer{
		logger: opts.Logger,
		queue:  make(chan []byte, opts.QueueCap),
	}

	writer.transport, err = transport.New(opts.transportConfig())
	if err != nil {
		return nil, err
	}

	writer.wg.Add(1)
	go writer.worker()

	return writer, nil
}

type Writer struct {
	transport transport.Transport

	queue  chan []byte
	logger Logger
	wg     sync.WaitGroup
}

// Write writes the data to the queue if it is not full.
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.write(append([]byte{}, p...))
}

func (w *Writer) write(p []byte) (n int, err error) {
	select {
	case w.queue <- p:
		return len(p), nil
	default:
		return 0, ErrWriteFailed
	}
}

// Close flushes any buffered log entries.
func (w *Writer) Close() error {
	close(w.queue)
	w.wg.Wait()

	return nil
}

func (w *Writer) worker() {
	for data := range w.queue {
		if !w.transport.IsConnected() {
			<-w.transport.IsReconnected()
		}

		w.wg.Add(1)
		go w.send(data)
	}

	w.wg.Done()
}

func (w *Writer) send(data []byte) {
	defer w.wg.Done()

	err := w.transport.Send(data)
	if err == nil {
		return
	}

	if w.logger != nil {
		w.logger.Printf("[error] send data failed: %v", err)
	}

	// if sending failed, return data to queue
	_, err = w.write(data)
	if err == nil {
		return
	}

	if w.logger != nil {
		w.logger.Printf("[error] %v", err)
	}
}

func processUrlString(url string) []string {
	urls := strings.Split(url, ",")

	for idx, val := range urls {
		urls[idx] = strings.TrimSpace(val)
	}

	return urls
}
