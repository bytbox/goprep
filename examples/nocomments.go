/* goprep demo : nocomments : removes all comments */

package main

import (
	"github.com/bytbox/goprep"
	"go/token"
)

func main() {
	tokIn, tokOut, done := goprep.StdInit()
	for tok := range tokIn {
		if tok.Token != token.COMMENT {
			str := tok.Str
			tokOut <- str
		}
	}
	close(tokOut)
	<-done
}
