package main

import (
	"flag"
	"fmt"
	"os"

	. "github.com/bytbox/goprep/goprep"
)

func main() {
	flag.Usage = func() {
		fmt.Println("usage: gppc < input.go > output.go")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		flag.Usage()
	}

	p := PipeInit(os.Stdin)
	Lines(p)
	Pass(True)(p)
	Discard(p)
	PipeEnd(p, os.Stdout)
}
