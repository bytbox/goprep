package main

import (
	"github.com/bytbox/goprep"
)

func main() {
	tokIn, tokOut, done := goprep.StdInit()
	for tok := range tokIn {
		str := tok.Str
		tokOut <- str
	}
	close(tokOut)
	<-done
}
