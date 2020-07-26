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

func (w *Writer) Write(p []byte) (n int, err error) {
	select {
	case w.queue <- append([]byte{}, p...):
		return len(p), nil
	default:
		return 0, ErrWriteFailed
	}
}

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
	if err := w.transport.Send(data); err != nil && w.logger != nil {
		w.logger.Printf("[error] send data failed: %v", err)
	}

	w.wg.Done()
}

// Process the url string argument to Connect.
// Return an array of urls, even if only one.
func processUrlString(url string) []string {
	urls := strings.Split(url, ",")
	for i, s := range urls {
		urls[i] = strings.TrimSpace(s)
	}
	return urls
}
