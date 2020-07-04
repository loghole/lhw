package main

import (
	"fmt"
	"time"

	"github.com/gadavy/lhw"
)

func main() {
	writer, err := lhw.NewWriter(lhw.Config{
		NodeURIs:    []string{"127.0.0.1:50000"},
		DropStorage: true,
	})
	if err != nil {
		panic(err)
	}

	defer writer.Close() // flushes storage, if any

	msg := fmt.Sprintf(`{"message": "test message", "time": %d}`, time.Now().UnixNano())

	_, err = writer.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}
