/* goprep demo : nocomments : removes all comments */

package main

import (
	. "github.com/bytbox/goprep"
	"go/token"
)

func main() {
	tokIn, tokOut, done := StdInit()
	tokIn = IgnoreType(tokIn, tokOut, token.COMMENT)
	for tok := range tokIn {
		str := tok.Str
		tokOut <- str
	}
	close(tokOut)
	<-done
}
