package main

import (
	. "github.com/bytbox/goprep/goprep"
	"os"
)

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)
	Discard(p)
	PipeEnd(p, os.Stdout)
}
