/* goprep demo : id : changes nothing */

package main

import (
	"goprep"
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
