package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gadavy/lhw"
)

func main() {
	writer, err := lhw.NewWriter(
		lhw.NodeWithAuth("https://127.0.0.1:50000", "secret_token"),
		lhw.WithInsecure(), lhw.WithLogger(log.New(os.Stdout, "", log.Ldate)),
	)
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
