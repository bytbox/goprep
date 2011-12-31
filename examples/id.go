package main

import (
	"github.com/bytbox/goprep"
	"os"
)

func main() {
	tokIn := goprep.Read(os.Stdin)
	tokOut, done := goprep.Write(os.Stdout)
	for tok := range tokIn {
		str := tok.Str
		tokOut <- str
	}
	close(tokOut)
	<-done
}
