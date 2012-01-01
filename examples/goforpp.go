package main

import (
	. "github.com/bytbox/goprep/goprep"
	"os"
)

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)

	nIn := make(chan TokenInfo)
	// we pass everything through until we reach 'gofor'
	go func(in chan TokenInfo,
		tOut chan TokenInfo,
		out chan string,
		sync chan interface{}) {
		for tok := range in {
			if tok.Str != "gofor" {
				out <- tok.Str
				continue
			}
			sync <- nil
		}
		close(tOut)
	}(p.Input, nIn, p.Output, p.Sync)
	p.Input = nIn

	Discard(p)
	PipeEnd(p, os.Stdout)
}
