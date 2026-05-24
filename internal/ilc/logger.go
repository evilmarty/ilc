package ilc

import (
	"io"
	"log"
)

var logger = log.New(io.Discard, "DEBUG: ", log.Lshortfile)
