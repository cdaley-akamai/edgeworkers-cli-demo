package devserver

import (
	"io"
	"log"
)

func newTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
