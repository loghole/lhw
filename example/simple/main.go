package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gadavy/lhw"
)

func main() {
	// Init writer.
	writer, err := lhw.NewWriter("http://secret_token_1@127.0.0.1:50000",
		lhw.WithLogger(log.New(os.Stdout, "", log.Ldate)),
	)
	if err != nil {
		panic(err)
	}

	// Close flushes any buffered log entries.
	defer writer.Close()

	msg := fmt.Sprintf(`{"message": "test message", "time": "%s"}`, time.Now().Format(time.RFC3339Nano))

	_, err = writer.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}
